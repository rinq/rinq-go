package commandamqp

import (
	"context"
	"sync"

	"github.com/jmalloc/twelf/src/twelf"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/rinq/rinq-go/src/internal/command"
	"github.com/rinq/rinq-go/src/internal/revisions"
	"github.com/rinq/rinq-go/src/internal/service"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/amqputil"
	"github.com/streadway/amqp"
)

type server struct {
	service.Service
	sm *service.StateMachine

	peerID    ident.PeerID
	preFetch  uint
	revisions revisions.Store
	queues    *QueueSet
	channels  amqputil.ChannelPool
	logger    twelf.Logger
	tracer    opentracing.Tracer

	parentCtx context.Context // parent of all contexts passed to handlers
	cancelCtx func()          // cancels parentCtx when the server stops

	// state-machine data
	channel    *amqp.Channel      // channel used for consuming
	deliveries chan amqp.Delivery // incoming command requests
	amqpClosed chan *amqp.Error
	pending    uint // number of requests currently being handled

	mutex    sync.RWMutex                   // guards handlers so handler can be read in dispatch() goroutine
	handlers map[string]rinq.CommandHandler // map of namespace to handler
}

// newServer creates, starts and returns a new server.
func newServer(
	peerID ident.PeerID,
	preFetch uint,
	revs revisions.Store,
	queues *QueueSet,
	channels amqputil.ChannelPool,
	logger twelf.Logger,
	tracer opentracing.Tracer,
) (command.Server, error) {
	s := &server{
		peerID:    peerID,
		preFetch:  preFetch,
		revisions: revs,
		queues:    queues,
		channels:  channels,
		logger:    logger,
		tracer:    tracer,

		deliveries: make(chan amqp.Delivery, preFetch),
		amqpClosed: make(chan *amqp.Error, 1),

		handlers: map[string]rinq.CommandHandler{},
	}

	s.sm = service.NewStateMachine(s.run, s.finalize)
	s.Service = s.sm

	if err := s.initialize(); err != nil {
		return nil, err
	}

	go s.sm.Run()

	return s, nil
}

func (s *server) Listen(ns string, h rinq.CommandHandler) (added bool, err error) {
	err = s.sm.Do(func() error {
		s.mutex.Lock()
		defer s.mutex.Unlock()

		if _, ok := s.handlers[ns]; ok {
			s.handlers[ns] = h
			return nil
		}

		s.handlers[ns] = h
		added = true

		return s.bind(ns)
	})

	return
}

func (s *server) Unlisten(ns string) (removed bool, err error) {
	err = s.sm.Do(func() error {
		s.mutex.Lock()
		defer s.mutex.Unlock()

		if _, ok := s.handlers[ns]; !ok {
			return nil
		}

		removed = true
		delete(s.handlers, ns)

		return s.unbind(ns)
	})

	return
}

func (s *server) bind(ns string) error {
	if err := s.channel.QueueBind(
		requestQueue(s.peerID),
		ns,
		multicastExchange,
		false, // noWait
		nil,   //  args
	); err != nil {
		return err
	}

	queue, err := s.queues.Get(ns)
	if err != nil {
		return err
	}

	messages, err := s.channel.Consume(
		queue,
		queue, // use queue name as consumer tag
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return err
	}

	go s.pipe(messages)

	return nil
}

func (s *server) unbind(ns string) error {
	if err := s.channel.QueueUnbind(
		requestQueue(s.peerID),
		ns,
		multicastExchange,
		nil, //  args
	); err != nil {
		return err
	}

	if err := s.channel.Cancel(
		balancedRequestQueue(ns), // use queue name as consumer tag
		false, // noWait
	); err != nil {
		return err
	}

	return s.queues.DeleteIfUnused(ns)
}

// initialize prepares the AMQP channel
func (s *server) initialize() error {
	if channel, err := s.channels.GetQOS(s.preFetch); err == nil { // do not return to pool, used for consume
		s.channel = channel
	} else {
		return err
	}

	s.channel.NotifyClose(s.amqpClosed)

	queue := requestQueue(s.peerID)

	if _, err := s.channel.QueueDeclare(
		queue,
		false, // durable
		false, // autoDelete
		true,  // exclusive,
		false, // noWait
		amqp.Table{"x-max-priority": priorityCount},
	); err != nil {
		return err
	}

	if err := s.channel.QueueBind(
		queue,
		s.peerID.String(),
		unicastExchange,
		false, // noWait
		nil,   // args
	); err != nil {
		return err
	}

	messages, err := s.channel.Consume(
		queue,
		queue, // use queue name as consumer tag
		false, // autoAck
		true,  // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return err
	}

	go s.pipe(messages)

	return nil
}

// run is the state entered when the service starts
func (s *server) run() (service.State, error) {
	logServerStart(s.logger, s.peerID, s.preFetch)

	s.parentCtx, s.cancelCtx = context.WithCancel(context.Background())

	for {
		select {
		case msg := <-s.deliveries:
			s.pending++
			go s.dispatch(&msg)

		case req := <-s.sm.Commands:
			s.sm.Execute(req)

		case <-s.sm.Graceful:
			return s.gracefulStopConsuming, nil

		case <-s.sm.Forceful:
			return nil, nil

		case err := <-s.amqpClosed:
			return nil, err
		}
	}
}

// gracefulStopConsuming is the first state entered when a graceful stop is
// requested.
func (s *server) gracefulStopConsuming() (service.State, error) {
	logServerStopping(s.logger, s.peerID, s.pending)

	queue := requestQueue(s.peerID)

	if err := s.channel.QueueUnbind(
		queue,
		s.peerID.String(),
		unicastExchange,
		nil, // args
	); err != nil {
		return nil, err
	}

	if err := s.channel.Cancel(
		queue, // use queue name as consumer tag
		false, // noWait
	); err != nil {
		return nil, err
	}

	// stop consuming from all namespace-based queues
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for ns := range s.handlers {
		if err := s.unbind(ns); err != nil {
			return nil, err
		}
	}

	return s.waitForHandlers, nil
}

// waitForHandlers is the second phase of a graceful stop. It waits for any
// pending command handlers to complete, while also rejecting any messages
// that have already been delivered.
func (s *server) waitForHandlers() (service.State, error) {
	for s.pending > 0 {
		select {
		case msg := <-s.deliveries:
			if err := msg.Reject(msg.Exchange == multicastExchange); err != nil { // (expr) = requeue
				return nil, err
			}

		case req := <-s.sm.Commands:
			s.sm.Execute(req)

		case <-s.sm.Forceful:
			return nil, nil
		}
	}

	return nil, nil
}

// finalize is the state-machine finalizer, it is called immediately before the
// Done() channel is closed.
func (s *server) finalize(err error) error {
	s.cancelCtx()
	logServerStop(s.logger, s.peerID, err)

	closeErr := s.channel.Close()

	// only report the closeErr if there's no causal error.
	if err == nil {
		return closeErr
	}

	return err
}

// dispatch validates an incoming command request and dispatches it the
// appropriate handler.
func (s *server) dispatch(msg *amqp.Delivery) {
	defer s.sm.DoGraceful(func() error {
		s.pending--
		return nil
	})

	// validate message ID
	msgID, err := ident.ParseMessageID(msg.MessageId)
	if err != nil {
		_ = msg.Reject(false) // false = don't requeue
		logServerInvalidMessageID(s.logger, s.peerID, msg.MessageId)
		return
	}

	// determine namespace + command
	ns, cmd, err := unpackNamespaceAndCommand(msg)
	if err != nil {
		_ = msg.Reject(false) // false = don't requeue
		logIgnoredMessage(s.logger, s.peerID, msgID, err)
		return
	}

	spanOpts, err := unpackSpanOptions(msg, s.tracer, ext.SpanKindRPCServer)
	if err != nil {
		_ = msg.Reject(false) // false = don't requeue
		logIgnoredMessage(s.logger, s.peerID, msgID, err)
		return
	}

	// find the handler for this namespace
	s.mutex.RLock()
	h, ok := s.handlers[ns]
	s.mutex.RUnlock()
	if !ok {
		_ = msg.Reject(msg.Exchange == balancedExchange) // requeue if "balanced"
		logNoLongerListening(s.logger, s.peerID, msgID, ns)
		return
	}

	// find the source session revision
	source, err := s.revisions.GetRevision(msgID.Ref)
	if err != nil {
		_ = msg.Reject(false) // false = don't requeue
		logIgnoredMessage(s.logger, s.peerID, msgID, err)
		return
	}

	s.handle(msgID, msg, ns, cmd, source, h, spanOpts)
}

// handle invokes the command handler for request.
func (s *server) handle(
	msgID ident.MessageID,
	msg *amqp.Delivery,
	ns string,
	cmd string,
	source rinq.Revision,
	handler rinq.CommandHandler,
	spanOpts []opentracing.StartSpanOption,
) {
	ctx := amqputil.UnpackTrace(s.parentCtx, msg)
	ctx, cancel := amqputil.UnpackDeadline(ctx, msg)
	defer cancel()

	span := s.tracer.StartSpan("", spanOpts...)
	defer span.Finish()

	ctx = opentracing.ContextWithSpan(ctx, span)

	req := rinq.Request{
		ID:        msgID,
		Source:    source,
		Namespace: ns,
		Command:   cmd,
		Payload:   rinq.NewPayloadFromBytes(msg.Body),
	}

	res, finalize := newResponse(
		ctx,
		s.channels,
		req,
		unpackReplyMode(msg),
	)

	if s.logger.IsDebug() {
		res = newDebugResponse(res)
		logRequestBegin(ctx, s.logger, s.peerID, msgID, req)
	}

	handler(ctx, req, res)

	if finalize() {
		_ = msg.Ack(false) // false = single message

		if dr, ok := res.(*debugResponse); ok {
			defer dr.Payload.Close()
			logRequestEnd(ctx, s.logger, s.peerID, msgID, req, dr.Payload, dr.Err)
		}
	} else if msg.Exchange == balancedExchange {
		select {
		case <-ctx.Done():
			_ = msg.Reject(false) // false = don't requeue
			logRequestRejected(ctx, s.logger, s.peerID, msgID, req, ctx.Err().Error())
		default:
			_ = msg.Reject(true) // true = requeue
			logRequestRequeued(ctx, s.logger, s.peerID, msgID, req)
		}
	} else {
		_ = msg.Reject(false) // false = don't requeue
		logRequestRejected(ctx, s.logger, s.peerID, msgID, req, "handler did not respond")
	}
}

// pipe aggregates AMQP messages from multiple consumers to a single channel.
func (s *server) pipe(messages <-chan amqp.Delivery) {
	for msg := range messages {
		select {
		case s.deliveries <- msg:
		case <-s.sm.Finalized:
		}
	}
}
