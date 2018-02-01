package rinqamqp

import (
	"context"
	"sync/atomic"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/internal/command"
	"github.com/rinq/rinq-go/src/internal/localsession"
	"github.com/rinq/rinq-go/src/internal/namespaces"
	"github.com/rinq/rinq-go/src/internal/notify"
	"github.com/rinq/rinq-go/src/internal/opentr"
	"github.com/rinq/rinq-go/src/internal/remotesession"
	"github.com/rinq/rinq-go/src/internal/service"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/trace"
	"github.com/streadway/amqp"
)

// peer is an AMQP-based implementation of rinq.Peer.
type peer struct {
	service.Service
	sm *service.StateMachine

	id          ident.PeerID
	broker      *amqp.Connection
	localStore  *localsession.Store
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
	localStore *localsession.Store,
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

	sess := localsession.NewSession(
		id,
		p.invoker,
		p.notifier,
		p.listener,
		p.logger,
		p.tracer,
	)

	p.localStore.Add(sess)
	go func() {
		<-sess.Done()
		p.localStore.Remove(sess.ID())
	}()

	return sess
}

func (p *peer) Listen(ns string, handler rinq.CommandHandler) error {
	namespaces.MustValidate(ns)

	added, err := p.server.Listen(
		ns,
		func(
			ctx context.Context,
			req rinq.Request,
			res rinq.Response,
		) {
			span := opentracing.SpanFromContext(ctx)

			traceID := trace.Get(ctx)

			opentr.SetupCommand(
				span,
				req.ID,
				req.Namespace,
				req.Command,
			)
			opentr.AddTraceID(span, traceID)
			opentr.LogServerRequest(span, p.id, req.Payload)

			handler(
				ctx,
				req,
				command.NewResponse(
					req,
					res,
					p.id,
					traceID,
					p.logger,
					span,
				),
			)
		},
	)

	if added {
		logStartedListening(p.logger, p.id, ns)
	}

	return err
}

func (p *peer) Unlisten(ns string) error {
	namespaces.MustValidate(ns)

	removed, err := p.server.Unlisten(ns)

	if removed {
		logStoppedListening(p.logger, p.id, ns)
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

	p.localStore.Each(func(sess *localsession.Session) {
		sess.Destroy()
		<-sess.Done()
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
