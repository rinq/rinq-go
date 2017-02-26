package amqp

import (
	"context"
	"sync/atomic"

	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpass/internal/command"
	"github.com/over-pass/overpass-go/src/overpass/internal/localsession"
	"github.com/over-pass/overpass-go/src/overpass/internal/notify"
	"github.com/over-pass/overpass-go/src/overpass/internal/remotesession"
	"github.com/over-pass/overpass-go/src/overpass/internal/service"
	"github.com/over-pass/overpass-go/src/overpass/internal/syncutil"
	"github.com/over-pass/overpass-go/src/overpass/internal/trace"
	"github.com/streadway/amqp"
)

// peer is an AMQP-based implementation of overpass.Peer.
type peer struct {
	service.Service
	sm *service.StateMachine

	id          overpass.PeerID
	broker      *amqp.Connection
	localStore  localsession.Store
	remoteStore remotesession.Store
	invoker     command.Invoker
	server      command.Server
	notifier    notify.Notifier
	listener    notify.Listener
	logger      overpass.Logger
	seq         uint32

	amqpClosed chan *amqp.Error
}

func newPeer(
	id overpass.PeerID,
	broker *amqp.Connection,
	localStore localsession.Store,
	remoteStore remotesession.Store,
	invoker command.Invoker,
	server command.Server,
	notifier notify.Notifier,
	listener notify.Listener,
	logger overpass.Logger,
) *peer {
	p := &peer{
		id:          id,
		broker:      broker,
		localStore:  localStore,
		remoteStore: remoteStore,
		invoker:     invoker,
		server:      server,
		notifier:    notifier,
		listener:    listener,
		logger:      logger,

		amqpClosed: make(chan *amqp.Error, 1),
	}

	p.sm = service.NewStateMachine(p.run, p.finalize)
	p.Service = p.sm

	broker.NotifyClose(p.amqpClosed)

	go p.sm.Run()

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
			req overpass.Request,
			res overpass.Response,
		) {
			handler(
				ctx,
				req,
				newLoggingResponse(
					req,
					res,
					p.id,
					trace.Get(ctx),
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

func (p *peer) run() (service.State, error) {
	select {
	case <-p.remoteStore.Done():
		return nil, p.remoteStore.Err()

	case <-p.invoker.Done():
		return nil, p.invoker.Err()

	case <-p.server.Done():
		return nil, p.server.Err()

	case <-p.listener.Done():
		return nil, p.listener.Err()

	case <-p.sm.Graceful:
		return p.graceful, nil

	case <-p.sm.Forceful:
		return nil, nil

	case err := <-p.amqpClosed:
		return nil, err
	}
}

func (p *peer) graceful() (service.State, error) {
	p.server.GracefulStop()
	p.invoker.GracefulStop()
	p.remoteStore.GracefulStop()
	p.listener.GracefulStop()

	done := syncutil.Group(
		p.remoteStore.Done(),
		p.invoker.Done(),
		p.server.Done(),
		p.listener.Done(),
	)

	select {
	case <-done:
		return nil, nil

	case <-p.sm.Forceful:
		return nil, nil

	case err := <-p.amqpClosed:
		return nil, err
	}
}

func (p *peer) finalize(err error) error {
	p.server.Stop()
	p.invoker.Stop()
	p.remoteStore.Stop()
	p.listener.Stop()

	p.localStore.Each(func(sess overpass.Session, _ localsession.Catalog) {
		sess.Close()
	})

	<-syncutil.Group(
		p.remoteStore.Done(),
		p.invoker.Done(),
		p.server.Done(),
		p.listener.Done(),
	)

	closeErr := p.broker.Close()

	// only return the close err if there's no causal error.
	if err == nil {
		return closeErr
	}

	return err
}
