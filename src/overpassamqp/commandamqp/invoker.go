package commandamqp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/over-pass/overpass-go/src/internals/command"
	"github.com/over-pass/overpass-go/src/internals/service"
	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/streadway/amqp"
)

// invoker is an AMQP-based implementation of command.Invoker
type invoker struct {
	service.Service
	closer *service.Closer

	peerID         overpass.PeerID
	preFetch       int
	defaultTimeout time.Duration
	queues         *queueSet
	channels       amqputil.ChannelPool
	channel        *amqp.Channel
	logger         overpass.Logger
	waiter         sync.WaitGroup

	mutex   sync.RWMutex
	pending map[string]chan returnValue
}

// newInvoker creates, initializes and returns a new invoker.
func newInvoker(
	peerID overpass.PeerID,
	preFetch int,
	defaultTimeout time.Duration,
	queues *queueSet,
	channels amqputil.ChannelPool,
	logger overpass.Logger,
) (command.Invoker, error) {
	svc, closer := service.NewImpl()

	i := &invoker{
		Service: svc,
		closer:  closer,

		peerID:         peerID,
		preFetch:       preFetch,
		defaultTimeout: defaultTimeout,
		queues:         queues,
		channels:       channels,
		logger:         logger,

		pending: map[string]chan returnValue{},
	}

	if err := i.initialize(); err != nil {
		return nil, err
	}

	return i, nil
}

// returnValue transports call response information across a pending call
// channel.
type returnValue struct {
	Payload *overpass.Payload
	Error   error
}

func (i *invoker) CallUnicast(
	ctx context.Context,
	msgID overpass.MessageID,
	target overpass.PeerID,
	namespace string,
	command string,
	payload *overpass.Payload,
) (string, *overpass.Payload, error) {
	return i.call(
		ctx,
		&amqp.Publishing{
			MessageId: msgID.String(),
			Priority:  callUnicastPriority,
			Type:      command,
			Headers:   amqp.Table{namespaceHeader: namespace},
			Body:      payload.Bytes(),
		},
		unicastExchange,
		target.String(),
		command,
		payload,
	)
}

func (i *invoker) CallBalanced(
	ctx context.Context,
	msgID overpass.MessageID,
	namespace string,
	command string,
	payload *overpass.Payload,
) (string, *overpass.Payload, error) {
	return i.call(
		ctx,
		&amqp.Publishing{
			MessageId: msgID.String(),
			Priority:  callBalancedPriority,
			Type:      command,
			Body:      payload.Bytes(),
		},
		balancedExchange,
		namespace,
		command,
		payload,
	)
}

func (i *invoker) ExecuteBalanced(
	ctx context.Context,
	msgID overpass.MessageID,
	namespace string,
	command string,
	payload *overpass.Payload,
) (string, error) {
	msg := amqp.Publishing{
		MessageId:    msgID.String(),
		Priority:     executePriority,
		Type:         command,
		DeliveryMode: amqp.Persistent,
		Body:         payload.Bytes(),
	}
	corrID := amqputil.PutCorrelationID(ctx, &msg)

	if err := i.send(balancedExchange, namespace, msg); err != nil {
		return corrID, err
	}

	return corrID, nil
}

func (i *invoker) ExecuteMulticast(
	ctx context.Context,
	msgID overpass.MessageID,
	namespace string,
	command string,
	payload *overpass.Payload,
) (string, error) {
	msg := amqp.Publishing{
		MessageId: msgID.String(),
		Priority:  executePriority,
		Type:      command,
		Body:      payload.Bytes(),
	}
	corrID := amqputil.PutCorrelationID(ctx, &msg)

	if err := i.send(multicastExchange, namespace, msg); err != nil {
		return corrID, err
	}

	return corrID, nil
}

func (i *invoker) initialize() error {
	channel, err := i.channels.Get() // do not return to pool, used for consume
	if err != nil {
		return err
	}

	if err = channel.Qos(i.preFetch, 0, true); err != nil {
		return err
	}

	queue := responseQueue(i.peerID)

	if _, err = channel.QueueDeclare(
		queue,
		false, // durable
		false, // autoDelete
		true,  // exclusive,
		false, // noWait
		nil,   // args
	); err != nil {
		return err
	}

	if err = channel.QueueBind(
		queue,
		i.peerID.String()+".*",
		responseExchange,
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

	i.channel = channel

	go i.monitor()
	go i.dispatchEach(messages)

	return nil
}

func (i *invoker) monitor() {
	logInvokerStart(i.logger, i.peerID, i.preFetch)

	var err error

	select {
	case err = <-i.channel.NotifyClose(make(chan *amqp.Error)):
	case <-i.closer.Stop():
		if i.closer.IsGraceful() {
			logInvokerStopping(i.logger, i.peerID)
			i.waiter.Wait()
		}
		i.channel.Close()
	}

	i.closer.Close(err)

	logInvokerStop(i.logger, i.peerID, err)
}

func (i *invoker) dispatchEach(messages <-chan amqp.Delivery) {
	for msg := range messages {
		msg.Ack(false)
		i.dispatch(msg)
	}
}

func (i *invoker) dispatch(msg amqp.Delivery) {
	i.mutex.RLock()
	channel := i.pending[msg.RoutingKey]
	i.mutex.RUnlock()

	if channel == nil {
		return // call has already reached deadline or been cancelled
	}

	var response returnValue

	switch msg.Type {
	case successResponse:
		response.Payload = overpass.NewPayloadFromBytes(msg.Body)

	case failureResponse:
		failureType, _ := msg.Headers[failureTypeHeader].(string)
		message, _ := msg.Headers[failureMessageHeader].(string)

		if failureType == "" {
			response.Error = errors.New("malformed response, failure type is empty")
		} else {
			response.Payload = overpass.NewPayloadFromBytes(msg.Body)
			response.Error = overpass.Failure{
				Type:    failureType,
				Message: message,
				Payload: response.Payload,
			}
		}

	case errorResponse:
		response.Error = overpass.UnexpectedError(msg.Body)

	default:
		response.Error = fmt.Errorf(
			"malformed response, message type '%s' is unexpected",
			msg.Type,
		)
	}

	channel <- response
	close(channel)
}

func (i *invoker) call(
	ctx context.Context,
	msg *amqp.Publishing,
	exchange string,
	key string,
	command string,
	payload *overpass.Payload,
) (
	string,
	*overpass.Payload,
	error,
) {
	i.waiter.Add(1)
	defer i.waiter.Done()

	msg.ReplyTo = "Y"
	corrID := amqputil.PutCorrelationID(ctx, msg)

	var cancel func()
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, i.defaultTimeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	if _, err := amqputil.PutExpiration(ctx, msg); err != nil {
		return corrID, nil, err
	}

	done := make(chan returnValue, 1)

	i.mutex.Lock()
	i.pending[msg.MessageId] = done
	i.mutex.Unlock()

	defer func() {
		i.mutex.Lock()
		delete(i.pending, msg.MessageId)
		i.mutex.Unlock()
	}()

	if err := i.send(exchange, key, *msg); err != nil {
		return corrID, nil, err
	}

	stop := i.closer.Stop()
	for {
		select {
		case <-ctx.Done():
			return corrID, nil, ctx.Err()
		case <-stop:
			stop = nil
			if !i.closer.IsGraceful() {
				cancel()
			}
		case response := <-done:
			return corrID, response.Payload, response.Error
		}
	}
}

func (i *invoker) send(exchange, key string, msg amqp.Publishing) error {
	channel, err := i.channels.Get()
	if err != nil {
		return err
	}
	defer i.channels.Put(channel)

	if exchange == balancedExchange {
		if _, err := i.queues.Get(channel, key); err != nil {
			return err
		}
	}

	return channel.Publish(
		exchange,
		key,
		false, // mandatory
		false, // immediate
		msg,
	)
}
