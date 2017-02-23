package commandamqp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/over-pass/overpass-go/src/internals/command"
	"github.com/over-pass/overpass-go/src/internals/deferutil"
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
		pending:        map[string]chan returnValue{},
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
	msg := amqp.Publishing{
		MessageId: msgID.String(),
		Priority:  callUnicastPriority,
		Type:      command,
		Headers:   amqp.Table{namespaceHeader: namespace},
		Body:      payload.Bytes(),
	}

	corrID, done, cancel, err := i.call(
		&ctx,
		&msg,
		unicastExchange,
		target.String(),
		command,
		payload,
	)
	if err != nil {
		return corrID, nil, err
	}
	defer cancel()

	select {
	case <-ctx.Done():
		return corrID, nil, ctx.Err()
	case response := <-done:
		return corrID, response.Payload, response.Error
	}
}

func (i *invoker) CallBalanced(
	ctx context.Context,
	msgID overpass.MessageID,
	namespace string,
	command string,
	payload *overpass.Payload,
) (string, *overpass.Payload, error) {
	msg := amqp.Publishing{
		MessageId: msgID.String(),
		Priority:  callBalancedPriority,
		Type:      command,
		Body:      payload.Bytes(),
	}

	corrID, done, cancel, err := i.call(
		&ctx,
		&msg,
		balancedExchange,
		namespace,
		command,
		payload,
	)
	if err != nil {
		return corrID, nil, err
	}
	defer cancel()

	select {
	case <-ctx.Done():
		return corrID, nil, ctx.Err()
	case response := <-done:
		return corrID, response.Payload, response.Error
	}
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

func (i *invoker) call(
	ctx *context.Context,
	msg *amqp.Publishing,
	exchange string,
	key string,
	command string,
	payload *overpass.Payload,
) (
	string,
	chan returnValue,
	func(),
	error,
) {
	cancel := deferutil.Set{}
	defer cancel.Run()

	if _, ok := (*ctx).Deadline(); !ok {
		timeoutCtx, cancelTimeout := context.WithTimeout(*ctx, i.defaultTimeout)
		*ctx = timeoutCtx
		cancel.Add(cancelTimeout)
	}

	if _, err := amqputil.PutExpiration(*ctx, msg); err != nil {
		return "", nil, nil, err
	}

	msg.ReplyTo = "Y"
	corrID := amqputil.PutCorrelationID(*ctx, msg)

	done := make(chan returnValue, 1)

	deferutil.With(&i.mutex, func() {
		i.pending[msg.MessageId] = done
	})

	cancel.Add(func() {
		i.mutex.Lock()
		defer i.mutex.Unlock()
		delete(i.pending, msg.MessageId)
	})

	if err := i.send(exchange, key, *msg); err != nil {
		return "", nil, nil, err
	}

	return corrID, done, cancel.Detach(), nil
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
	go i.consume(messages)

	return nil
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

func (i *invoker) consume(messages <-chan amqp.Delivery) {
	closed := i.channel.NotifyClose(make(chan *amqp.Error))

	for {
		select {
		case err := <-closed:
			i.closer.Close(err)
			// TODO: log
			return

		case <-i.closer.Stop():
			i.channel.Close()
			i.closer.Close(nil)
			// TODO: log
			return

		case msg, ok := <-messages:
			if ok {
				msg.Ack(false)
				i.dispatch(msg)
			} else {
				messages = nil
			}
		}
	}
}

func (i *invoker) dispatch(msg amqp.Delivery) {
	var channel chan returnValue
	deferutil.RWith(&i.mutex, func() {
		channel = i.pending[msg.RoutingKey]
	})

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
			if message == "" {
				message = "unknown error"
			}

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
}
