package commandamqp

import (
	"context"
	"sync"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/amqp/internal/amqputil"
	"github.com/rinq/rinq-go/src/rinq/internal/command"
	"github.com/rinq/rinq-go/src/rinq/internal/revision"
	"github.com/rinq/rinq-go/src/rinq/internal/service"
	"github.com/streadway/amqp"
)

type server struct {
	service.Service
	sm *service.StateMachine

	peerID    rinq.PeerID
	preFetch  int
	revisions revision.Store
	queues    *queueSet
	channels  amqputil.ChannelPool
	logger    rinq.Logger

	parentCtx context.Context // parent of all contexts passed to handlers
	cancelCtx func()          // cancels parentCtx when the server stops

	mutex    sync.RWMutex
	channel  *amqp.Channel                  // channel used for consuming
	handlers map[string]rinq.CommandHandler // map of namespace to handler

	deliveries chan amqp.Delivery // incoming command requests
	handled    chan struct{}      // signals requests have been handled
	amqpClosed chan *amqp.Error

	// state-machine data
	pending uint // number of requests currently being handled
}

// newServer creates, starts and returns a new server.
func newServer(
	peerID rinq.PeerID,
	preFetch int,
	revisions revision.Store,
	queues *queueSet,
	channels amqputil.ChannelPool,
	logger rinq.Logger,
) (command.Server, error) {
	s := &server{
		peerID:    peerID,
		preFetch:  preFetch,
		revisions: revisions,
		queues:    queues,
		channels:  channels,
		logger:    logger,

		handlers: map[string]rinq.CommandHandler{},

		deliveries: make(chan amqp.Delivery, preFetch),
		handled:    make(chan struct{}, preFetch),
		amqpClosed: make(chan *amqp.Error, 1),
	}

	s.sm = service.NewStateMachine(s.run, s.finalize)
	s.Service = s.sm

	if err := s.initialize(); err != nil {
		return nil, err
	}

	go s.sm.Run()

	return s, nil
}

func (s *server) Listen(namespace string, handler rinq.CommandHandler) (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// we're already listening, just swap the handler
	if _, ok := s.handlers[namespace]; ok {
		s.handlers[namespace] = handler
		return false, nil
	}

	if err := s.channel.QueueBind(
		requestQueue(s.peerID),
		namespace,
		multicastExchange,
		false, // noWait
		nil,   //  args
	); err != nil {
		return false, err
	}

	queue, err := s.queues.Get(s.channel, namespace)
	if err != nil {
		return false, err
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
		return false, err
	}

	s.handlers[namespace] = handler
	go s.pipe(messages)

	return true, nil
}

func (s *server) Unlisten(namespace string) (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.handlers[namespace]; !ok {
		return false, nil
	}

	if err := s.channel.QueueUnbind(
		requestQueue(s.peerID),
		namespace,
		multicastExchange,
		nil, //  args
	); err != nil {
		return false, err
	}

	if err := s.channel.Cancel(
		balancedRequestQueue(namespace), // use queue name as consumer tag
		false, // noWait
	); err != nil {
		return false, err
	}

	delete(s.handlers, namespace)

	return true, nil
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

		case <-s.handled:
			s.pending--

		case <-s.sm.Graceful:
			return s.graceful, nil

		case <-s.sm.Forceful:
			return s.forceful, nil

		case err := <-s.amqpClosed:
			return nil, err
		}
	}
}

// graceful is the state entered when a graceful stop is requested
func (s *server) graceful() (service.State, error) {
	logServerStopping(s.logger, s.peerID, s.pending)

	if err := s.closeChannel(); err != nil {
		return nil, err
	}

	for s.pending > 0 {
		select {
		case <-s.handled:
			s.pending--

		case <-s.sm.Forceful:
			return nil, nil
		}
	}

	return nil, nil
}

// forceful is the state entered when a stop is requested
func (s *server) forceful() (service.State, error) {
	return nil, s.closeChannel()
}

// finalize is the state-machine finalizer, it is called immediately before the
// Done() channel is closed.
func (s *server) finalize(err error) error {
	s.cancelCtx()
	logServerStop(s.logger, s.peerID, err)
	return err

}

// dispatch validates an incoming command request and dispatches it the
// appropriate handler.
func (s *server) dispatch(msg *amqp.Delivery) {
	defer func() {
		select {
		case s.handled <- struct{}{}:
		case <-s.sm.Forceful:
		}
	}()

	// validate message ID
	msgID, err := rinq.ParseMessageID(msg.MessageId)
	if err != nil {
		msg.Reject(false)
		logServerInvalidMessageID(s.logger, s.peerID, msg.MessageId)
		return
	}

	// determine namespace + command
	ns, cmd, err := unpackNamespaceAndCommand(msg)
	if err != nil {
		msg.Reject(false)
		logIgnoredMessage(s.logger, s.peerID, msgID, err)
		return
	}

	// find the handler for this namespace
	s.mutex.RLock()
	h, ok := s.handlers[ns]
	s.mutex.RUnlock()
	if !ok {
		msg.Reject(msg.Exchange == balancedExchange) // requeue if "balanced"
		logNoLongerListening(s.logger, s.peerID, msgID, ns)
		return
	}

	// find the source session revision
	source, err := s.revisions.GetRevision(msgID.Session)
	if err != nil {
		msg.Reject(false)
		logIgnoredMessage(s.logger, s.peerID, msgID, err)
		return
	}

	s.handle(msgID, msg, ns, cmd, source, h)
}

// handle invokes the command handler for request.
func (s *server) handle(
	msgID rinq.MessageID,
	msg *amqp.Delivery,
	ns string,
	cmd string,
	source rinq.Revision,
	handler rinq.CommandHandler,
) {
	ctx := amqputil.UnpackTrace(s.parentCtx, msg)
	ctx, cancel := amqputil.UnpackDeadline(ctx, msg)
	defer cancel()

	req := rinq.Request{
		Source:      source,
		Namespace:   ns,
		Command:     cmd,
		Payload:     rinq.NewPayloadFromBytes(msg.Body),
		IsMulticast: msg.Exchange == multicastExchange,
	}

	res, finalize := newResponse(
		ctx,
		s.channels,
		msgID,
		req,
		unpackReplyMode(msg),
	)

	if s.logger.IsDebug() {
		res = newDebugResponse(res)
		logRequestBegin(ctx, s.logger, s.peerID, msgID, req)
	}

	handler(ctx, req, res)

	if finalize() {
		msg.Ack(false) // false = single message

		if dr, ok := res.(*debugResponse); ok {
			defer dr.Payload.Close()
			logRequestEnd(ctx, s.logger, s.peerID, msgID, req, dr.Payload, dr.Err)
		}
	} else if msg.Exchange == balancedExchange {
		select {
		case <-ctx.Done():
			msg.Reject(false) // false = don't requeue
			logRequestRejected(ctx, s.logger, s.peerID, msgID, req, ctx.Err().Error())
		default:
			msg.Reject(true) // true = requeue
			logRequestRequeued(ctx, s.logger, s.peerID, msgID, req)
		}
	} else {
		msg.Reject(false) // false = don't requeue
		logRequestRejected(ctx, s.logger, s.peerID, msgID, req, ctx.Err().Error())
	}
}

// pipe aggregates AMQP messages from multiple consumers a single channel.
func (s *server) pipe(messages <-chan amqp.Delivery) {
	for msg := range messages {
		select {
		case s.deliveries <- msg:
		case <-s.sm.Finalized:
		}
	}
}

func (s *server) closeChannel() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.channel.Close()
}
