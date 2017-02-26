package commandamqp

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/over-pass/overpass-go/src/internal/command"
	"github.com/over-pass/overpass-go/src/internal/revision"
	"github.com/over-pass/overpass-go/src/internal/service"
	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpassamqp/internal/amqputil"
	"github.com/streadway/amqp"
)

type server struct {
	service.Service
	sm *service.StateMachine

	peerID    overpass.PeerID
	preFetch  int
	revisions revision.Store
	queues    *queueSet
	channels  amqputil.ChannelPool
	logger    overpass.Logger

	parentCtx context.Context // parent of all contexts passed to handlers
	cancelCtx func()          // cancels parentCtx when the server stops

	mutex    sync.RWMutex
	channel  *amqp.Channel                      // channel used for consuming
	handlers map[string]overpass.CommandHandler // map of namespace to handler

	deliveries chan amqp.Delivery // incoming command requests
	handled    chan struct{}      // signals requests have been handled
	amqpClosed chan *amqp.Error

	// state-machine data
	pending uint // number of requests currently being handled
}

// newServer creates, starts and returns a new server.
func newServer(
	peerID overpass.PeerID,
	preFetch int,
	revisions revision.Store,
	queues *queueSet,
	channels amqputil.ChannelPool,
	logger overpass.Logger,
) (command.Server, error) {
	s := &server{
		peerID:    peerID,
		preFetch:  preFetch,
		revisions: revisions,
		queues:    queues,
		channels:  channels,
		logger:    logger,

		handlers: map[string]overpass.CommandHandler{},

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

func (s *server) Listen(namespace string, handler overpass.CommandHandler) (bool, error) {
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
	if channel, err := s.channels.Get(); err == nil { // do not return to pool, used for consume
		s.channel = channel
	} else {
		return err
	}

	s.channel.NotifyClose(s.amqpClosed)

	if err := s.channel.Qos(s.preFetch, 0, true); err != nil {
		return err
	}

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

	if err := s.channel.Close(); err != nil {
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
	return nil, s.channel.Close()
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
	msgID, err := overpass.ParseMessageID(msg.MessageId)
	if err != nil {
		msg.Reject(false)
		logInvalidMessageID(s.logger, s.peerID, msg.MessageId)
		return
	}

	// determine command namespace
	namespace, err := s.unpackNamespace(msg)
	if err != nil {
		msg.Reject(false)
		logIgnoredMessage(s.logger, s.peerID, msgID, err)
		return
	}

	// find the handler for this namespace
	s.mutex.RLock()
	handler, ok := s.handlers[namespace]
	s.mutex.RUnlock()
	if !ok {
		msg.Reject(msg.Exchange == balancedExchange) // requeue if "balanced"
		logNoLongerListening(s.logger, s.peerID, msgID, namespace)
		return
	}

	// find the source session revision
	source, err := s.revisions.GetRevision(msgID.Session)
	if err != nil {
		msg.Reject(false)
		logIgnoredMessage(s.logger, s.peerID, msgID, err)
		return
	}

	s.handle(msgID, msg, namespace, source, handler)
}

// handle invokes the command handler for request.
func (s *server) handle(
	msgID overpass.MessageID,
	msg *amqp.Delivery,
	namespace string,
	source overpass.Revision,
	handler overpass.CommandHandler,
) {
	ctx := amqputil.UnpackTrace(s.parentCtx, msg)
	ctx, cancel := amqputil.UnpackDeadline(ctx, msg)
	defer cancel()

	cmd := overpass.Command{
		Source:      source,
		Namespace:   namespace,
		Command:     msg.Type,
		Payload:     overpass.NewPayloadFromBytes(msg.Body),
		IsMulticast: msg.Exchange == multicastExchange,
	}

	var res overpass.Responder = &responder{
		channels:   s.channels,
		context:    ctx,
		msgID:      msgID,
		isRequired: msg.ReplyTo != "",
	}

	if s.logger.IsDebug() {
		res = newCapturingResponder(res)
		logRequestBegin(ctx, s.logger, s.peerID, msgID, cmd)
	}

	handler(ctx, cmd, res)

	if res.IsClosed() {
		msg.Ack(false) // false = single message

		if s.logger.IsDebug() {
			cap := res.(*capturingResponder)
			payload, err := cap.Response()
			defer payload.Close()
			logRequestEnd(ctx, s.logger, s.peerID, msgID, cmd, payload, err)
		}
	} else if msg.Exchange == balancedExchange {
		select {
		case <-ctx.Done():
			msg.Reject(false) // false = don't requeue
			logRequestRejected(ctx, s.logger, s.peerID, msgID, cmd, ctx.Err().Error())
		default:
			msg.Reject(true) // true = requeue
			logRequestRequeued(ctx, s.logger, s.peerID, msgID, cmd)
		}
	} else {
		msg.Reject(false) // false = don't requeue
		logRequestRejected(ctx, s.logger, s.peerID, msgID, cmd, ctx.Err().Error())
	}
}

// unpackNamespace extracts and validates the namespace in a command request.
func (s *server) unpackNamespace(msg *amqp.Delivery) (string, error) {
	switch msg.Exchange {
	case balancedExchange, multicastExchange:
		return msg.RoutingKey, nil

	case unicastExchange:
		if namespace, ok := msg.Headers[namespaceHeader].(string); ok {
			return namespace, nil
		}

		return "", errors.New("malformed request, namespace is not a string")

	default:
		return "", fmt.Errorf("delivery via '%s' exchange is not expected", msg.Exchange)
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
