package overpassamqp

import (
	"context"
	"sync/atomic"

	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/over-pass/overpass-go/src/internals/command"
	"github.com/over-pass/overpass-go/src/internals/localsession"
	"github.com/over-pass/overpass-go/src/internals/notify"
	"github.com/over-pass/overpass-go/src/internals/remotesession"
	"github.com/over-pass/overpass-go/src/internals/service"
	"github.com/over-pass/overpass-go/src/overpass"
)

// peer is an AMQP-based implementation of overpass.Peer.
type peer struct {
	service.Service
	closer *service.Closer

	id          overpass.PeerID
	connection  service.Service
	localStore  localsession.Store
	remoteStore remotesession.Store
	invoker     command.Invoker
	server      command.Server
	notifier    notify.Notifier
	listener    notify.Listener
	logger      overpass.Logger
	seq         uint32
}

func newPeer(
	id overpass.PeerID,
	connection service.Service,
	localStore localsession.Store,
	remoteStore remotesession.Store,
	invoker command.Invoker,
	server command.Server,
	notifier notify.Notifier,
	listener notify.Listener,
	logger overpass.Logger,
) *peer {
	svc, closer := service.NewImpl()

	p := &peer{
		Service: svc,
		closer:  closer,

		id:          id,
		connection:  connection,
		localStore:  localStore,
		remoteStore: remoteStore,
		invoker:     invoker,
		server:      server,
		notifier:    notifier,
		listener:    listener,
		logger:      logger,
	}

	go p.monitor()

	return p
}

func (p *peer) ID() overpass.PeerID {
	return p.id
}

func (p *peer) Session() overpass.Session {
	id := overpass.SessionID{
		Peer: p.id,
		Seq:  atomic.AddUint32(&p.seq, 1),
	}

	cat := localsession.NewCatalog(id, p.logger)
	sess := localsession.NewSession(
		id,
		cat,
		p.invoker,
		p.notifier,
		p.listener,
		p.logger,
	)

	p.localStore.Add(sess, cat)
	go func() {
		<-sess.Done()
		p.localStore.Remove(sess.ID())
	}()

	return sess
}

func (p *peer) Listen(namespace string, handler overpass.CommandHandler) error {
	if err := overpass.ValidateNamespace(namespace); err != nil {
		return err
	}

	added, err := p.server.Listen(
		namespace,
		func(
			ctx context.Context,
			cmd overpass.Command,
			res overpass.Responder,
		) {
			handler(
				ctx,
				cmd,
				newLoggingResponder(
					res,
					p.id,
					amqputil.GetCorrelationID(ctx),
					cmd,
					p.logger,
				),
			)
		},
	)

	if added {
		logStartedListening(p.logger, p.id, namespace)
	}

	return err
}

func (p *peer) Unlisten(namespace string) error {
	if err := overpass.ValidateNamespace(namespace); err != nil {
		return err
	}

	removed, err := p.server.Unlisten(namespace)

	if removed {
		logStoppedListening(p.logger, p.id, namespace)
	}

	return err
}

func (p *peer) monitor() {
	var err error

	// wait for ANY of the services to stop
	select {
	case <-p.connection.Done():
		err = p.connection.Err()
	case <-p.remoteStore.Done():
		err = p.remoteStore.Err()
	case <-p.invoker.Done():
		err = p.invoker.Err()
	case <-p.server.Done():
		err = p.server.Err()
	case <-p.listener.Done():
		err = p.listener.Err()
	case <-p.closer.Stop():
	}

	services := []service.Service{
		p.server,
		p.invoker,
		p.remoteStore,
		p.listener,
	}

	// ask ALL services to stop
	for _, svc := range services {
		if p.closer.IsGraceful() {
			go svc.GracefulStop()
		} else {
			go svc.Stop()
		}
	}

	// wait for ALL services to stop
	for _, svc := range services {
		<-svc.Done()
	}

	p.localStore.Each(func(sess overpass.Session, _ localsession.Catalog) {
		sess.Close()
	})

	p.connection.Stop()
	p.closer.Close(err)
}
