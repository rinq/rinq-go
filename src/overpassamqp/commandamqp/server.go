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
	closer *service.Closer

	peerID    overpass.PeerID
	preFetch  int
	revisions revision.Store
	queues    *queueSet
	channels  amqputil.ChannelPool
	logger    overpass.Logger
	parentCtx context.Context
	cancelCtx func()
	waiter    sync.WaitGroup

	mutex    sync.RWMutex
	channel  *amqp.Channel
	handlers map[string]overpass.CommandHandler
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
	svc, closer := service.NewImpl()

	s := &server{
		Service: svc,
		closer:  closer,

		peerID:    peerID,
		preFetch:  preFetch,
		revisions: revisions,
		queues:    queues,
		channels:  channels,
		logger:    logger,
		handlers:  map[string]overpass.CommandHandler{},
	}

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
	go s.dispatchEach(messages)

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
	channel, err := s.channels.Get() // do not return to pool, used for consume
	if err != nil {
		return err
	}

	if err = channel.Qos(s.preFetch, 0, true); err != nil {
		return err
	}

	queue := requestQueue(s.peerID)

	if _, err = channel.QueueDeclare(
		queue,
		false, // durable
		false, // autoDelete
		true,  // exclusive,
		false, // noWait
		amqp.Table{"x-max-priority": priorityCount},
	); err != nil {
		return err
	}

	if err = channel.QueueBind(
		queue,
		s.peerID.String(),
		unicastExchange,
		false, // noWait
		nil,   // args
	); err != nil {
		return err
	}

	messages, err := channel.Consume(
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

	s.channel = channel

	go s.monitor()
	go s.dispatchEach(messages)

	return nil
}

func (s *server) monitor() {
	logServerStart(s.logger, s.peerID, s.preFetch)

	var err error

	select {
	case err = <-s.channel.NotifyClose(make(chan *amqp.Error)):
	case <-s.closer.Stop():
		s.channel.Close()
		if s.closer.IsGraceful() {
			logServerStopping(s.logger, s.peerID)
			s.waiter.Wait()
		}
	}

	s.cancelCtx()
	s.closer.Close(err)

	logServerStop(s.logger, s.peerID, err)
}

func (s *server) dispatchEach(messages <-chan amqp.Delivery) {
	for msg := range messages {
		s.waiter.Add(1)
		go s.dispatch(msg)
	}
}

func (s *server) dispatch(msg amqp.Delivery) {
	defer s.waiter.Done()

	msgID, err := overpass.ParseMessageID(msg.MessageId)
	if err != nil {
		msg.Reject(false)
		logInvalidMessageID(s.logger, s.peerID, msg)
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
}

func (s *server) handle(msgID overpass.MessageID, namespace string, msg amqp.Delivery) error {
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
	// TODO: defer invalidate responder

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
