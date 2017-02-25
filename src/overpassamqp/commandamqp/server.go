package commandamqp

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/over-pass/overpass-go/src/internals/command"
	"github.com/over-pass/overpass-go/src/internals/revision"
	"github.com/over-pass/overpass-go/src/internals/service"
	"github.com/over-pass/overpass-go/src/overpass"
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
	parentCtx context.Context
	cancelCtx func()

	mutex    sync.RWMutex
	channel  *amqp.Channel
	handlers map[string]overpass.CommandHandler

	deliveries  chan *amqp.Delivery
	handlerDone chan struct{}
	amqpClosed  chan *amqp.Error

	pending uint
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

		deliveries:  make(chan *amqp.Delivery),
		handlerDone: make(chan struct{}),
		amqpClosed:  make(chan *amqp.Error),
	}

	s.sm = service.NewStateMachine(s.run, s.finalize)
	s.Service = s.sm

	s.parentCtx, s.cancelCtx = context.WithCancel(context.Background())

	if err := s.initialize(); err != nil {
		return nil, err
	}

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

func (s *server) initialize() error {
	if channel, err := s.channels.Get(); err == nil { // do not return to pool, used for invoker
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

	go s.sm.Run()

	return nil
}

func (s *server) run() (service.State, error) {
	logServerStart(s.logger, s.peerID, s.preFetch)

	for {
		select {
		case msg := <-s.deliveries:
			s.pending++
			go s.dispatch(msg)

		case <-s.handlerDone:
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

func (s *server) graceful() (service.State, error) {
	logServerStopping(s.logger, s.peerID, s.pending)

	if err := s.channel.Close(); err != nil {
		return nil, err
	}

	for s.pending > 0 {
		select {
		case <-s.handlerDone:
			s.pending--

		case <-s.sm.Forceful:
			return nil, nil
		}
	}

	return nil, nil
}

func (s *server) forceful() (service.State, error) {
	return nil, s.channel.Close()
}

func (s *server) finalize(err error) error {
	l.cancelCtx()
	logServerStop(s.logger, s.peerID, err)
	return err

}

func (s *server) dispatch(msg *amqp.Delivery) {
	msgID, err := overpass.ParseMessageID(msg.MessageId)
	if err != nil {
		msg.Reject(false)
		logInvalidMessageID(s.logger, s.peerID, msg.MessageId)
		return
	}

	switch msg.Exchange {
	case balancedExchange, multicastExchange:
		err = s.handle(msgID, msg.RoutingKey, msg)
	case unicastExchange:
		if namespace, ok := msg.Headers[namespaceHeader].(string); ok {
			err = s.handle(msgID, namespace, msg)
		} else {
			err = errors.New("malformed request, namespace is not a string")
		}
	default:
		err = fmt.Errorf("delivery via '%s' exchange is not expected", msg.Exchange)
	}

	if err != nil {
		msg.Reject(false)
		logIgnoredMessage(s.logger, s.peerID, msgID, err)
	}

	s.handlerDone <- struct{}{}
}

func (s *server) handle(msgID overpass.MessageID, namespace string, msg *amqp.Delivery) error {
	s.mutex.RLock()
	handler := s.handlers[namespace]
	s.mutex.RUnlock()

	if handler == nil {
		msg.Reject(true)
		logNoLongerListening(s.logger, s.peerID, msgID, namespace)
		return nil
	}

	source, err := s.revisions.GetRevision(msgID.Session)
	if err != nil {
		return err
	}

	ctx := amqputil.WithCorrelationID(s.parentCtx, msg)
	ctx, cancel := amqputil.WithExpiration(ctx, msg)
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
		msg.Ack(false)

		if s.logger.IsDebug() {
			cap := res.(*capturingResponder)
			payload, err := cap.Response()
			defer payload.Close()
			logRequestEnd(ctx, s.logger, s.peerID, msgID, cmd, payload, err)
		}

		return nil
	}

	if msg.Exchange == balancedExchange {
		select {
		case <-ctx.Done():
		default:
			msg.Reject(true)
			logRequestRequeued(ctx, s.logger, s.peerID, msgID, cmd)
			return nil
		}
	}

	msg.Reject(false)
	logRequestRejected(ctx, s.logger, s.peerID, msgID, cmd)
	return nil
}

func (s *server) pipe(messages <-chan amqp.Delivery) {
	for msg := range messages {
		s.deliveries <- &msg
	}
}
