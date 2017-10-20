package amqp

import (
	"context"
	"sync/atomic"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/command"
	"github.com/rinq/rinq-go/src/rinq/internal/localsession"
	"github.com/rinq/rinq-go/src/rinq/internal/namespaces"
	"github.com/rinq/rinq-go/src/rinq/internal/notify"
	"github.com/rinq/rinq-go/src/rinq/internal/opentr"
	"github.com/rinq/rinq-go/src/rinq/internal/remotesession"
	"github.com/rinq/rinq-go/src/rinq/internal/service"
	"github.com/rinq/rinq-go/src/rinq/trace"
	"github.com/streadway/amqp"
)

// peer is an AMQP-based implementation of rinq.Peer.
type peer struct {
	service.Service
	sm *service.StateMachine

	id          ident.PeerID
	broker      *amqp.Connection
	localStore  localsession.Store
	remoteStore remotesession.Store
	invoker     command.Invoker
	server      command.Server
	notifier    notify.Notifier
	listener    notify.Listener
	logger      rinq.Logger
	tracer      opentracing.Tracer

	seq        uint32
	amqpClosed chan *amqp.Error
}

func newPeer(
	id ident.PeerID,
	broker *amqp.Connection,
	localStore localsession.Store,
	remoteStore remotesession.Store,
	invoker command.Invoker,
	server command.Server,
	notifier notify.Notifier,
	listener notify.Listener,
	logger rinq.Logger,
	tracer opentracing.Tracer,
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
		tracer:      tracer,

		amqpClosed: make(chan *amqp.Error, 1),
	}

	p.sm = service.NewStateMachine(p.run, p.finalize)
	p.Service = p.sm

	broker.NotifyClose(p.amqpClosed)

	go p.sm.Run()

	return p
}

func (p *peer) ID() ident.PeerID {
	return p.id
}

func (p *peer) Session() rinq.Session {
	id := p.id.Session(
		atomic.AddUint32(&p.seq, 1),
	)

	cat := localsession.NewCatalog(id, p.logger)
	sess := localsession.NewSession(
		id,
		cat,
		p.invoker,
		p.notifier,
		p.listener,
		p.logger,
		p.tracer,
	)

	p.localStore.Add(sess, cat)
	go func() {
		<-sess.Done()
		p.localStore.Remove(sess.ID())
	}()

	return sess
}

func (p *peer) Listen(namespace string, handler rinq.CommandHandler) error {
	if err := namespaces.Validate(namespace); err != nil {
		return err
	}

	added, err := p.server.Listen(
		namespace,
		func(
			ctx context.Context,
			req rinq.Request,
			res rinq.Response,
		) {
			span := opentracing.SpanFromContext(ctx)

			opentr.SetupCommand(
				span,
				req.ID,
				req.Namespace,
				req.Command,
			)
			opentr.LogServerRequest(span, p.id, req.Payload)

			handler(
				ctx,
				req,
				command.NewResponse(
					req,
					res,
					p.id,
					trace.Get(ctx),
					p.logger,
					span,
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
	if err := namespaces.Validate(namespace); err != nil {
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

	done := service.WaitAll(
		p.remoteStore,
		p.invoker,
		p.server,
		p.listener,
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

	p.localStore.Each(func(sess rinq.Session, _ localsession.Catalog) {
		sess.Destroy()
	})

	<-service.WaitAll(
		p.remoteStore,
		p.invoker,
		p.server,
		p.listener,
	)

	closeErr := p.broker.Close()

	// only return the close err if there's no causal error.
	if err == nil {
		return closeErr
	}

	return err
}
