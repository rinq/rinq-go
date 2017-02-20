package commandamqp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/over-pass/overpass-go/src/internals/command"
	"github.com/over-pass/overpass-go/src/internals/deferutil"
	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/streadway/amqp"
)

type invoker struct {
	peerID         overpass.PeerID
	defaultTimeout time.Duration
	queues         *queueSet
	channels       amqputil.ChannelPool
	logger         *log.Logger

	mutex   sync.RWMutex
	pending map[string]chan returnValue

	channel *amqp.Channel
	done    chan struct{}
	err     atomic.Value
}

// newInvoker creates, initializes and returns a new invoker.
func newInvoker(
	peerID overpass.PeerID,
	defaultTimeout time.Duration,
	queues *queueSet,
	channels amqputil.ChannelPool,
	logger *log.Logger,
) (command.Invoker, error) {
	i := &invoker{
		peerID:         peerID,
		defaultTimeout: defaultTimeout,
		queues:         queues,
		channels:       channels,
		logger:         logger,
		pending:        map[string]chan returnValue{},
		done:           make(chan struct{}),
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
) (*overpass.Payload, error) {
	msg := amqp.Publishing{
		MessageId: msgID.String(),
		Priority:  callPriority,
		Type:      command,
		Headers:   amqp.Table{namespaceHeader: namespace},
		Body:      payload.Bytes(),
	}

	_, done, cancel, err := i.call(
		&ctx,
		&msg,
		unicastExchange,
		target.String(),
		command,
		payload,
	)
	if err != nil {
		return nil, err
	}
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case response := <-done:
		return response.Payload, response.Error
	}
}

func (i *invoker) CallBalanced(
	ctx context.Context,
	msgID overpass.MessageID,
	namespace string,
	command string,
	payload *overpass.Payload,
) (*overpass.Payload, error) {
	msg := amqp.Publishing{
		MessageId: msgID.String(),
		Priority:  callPriority,
		Type:      command,
		Body:      payload.Bytes(),
	}

	sentAt := time.Now()
	corrID, done, cancel, err := i.call(
		&ctx,
		&msg,
		balancedExchange,
		namespace,
		command,
		payload,
	)
	if err != nil {
		return nil, err
	}
	defer cancel()

	var response returnValue
	var action string
	var info string

	select {
	case <-ctx.Done():
		response.Error = ctx.Err()
		action = "aborted"
		info = response.Error.Error()

	case response = <-done:
		if response.Error == nil {
			action = "returned"
			info = fmt.Sprintf("%d bytes", response.Payload.Len())
		} else {
			action = "failed"
			info = response.Error.Error()
		}
	}

	i.logger.Printf(
		"%s called '%s' in '%s' namespace (%d bytes), %s after %dms (%s) [%s]",
		msgID.ShortString(),
		command,
		namespace,
		payload.Len(),
		action,
		time.Now().Sub(sentAt)/time.Millisecond,
		info,
		corrID,
	)

	return response.Payload, response.Error
}

func (i *invoker) ExecuteBalanced(
	ctx context.Context,
	msgID overpass.MessageID,
	namespace string,
	command string,
	payload *overpass.Payload,
) error {
	msg := amqp.Publishing{
		MessageId:    msgID.String(),
		Priority:     executePriority,
		Type:         command,
		DeliveryMode: amqp.Persistent,
		Body:         payload.Bytes(),
	}
	corrID := amqputil.PutCorrelationID(ctx, &msg)

	if err := i.send(balancedExchange, namespace, msg); err != nil {
		return err
	}

	i.logger.Printf(
		"%s executed '%s' in '%s' namespace (%d bytes) [%s]",
		msgID.ShortString(),
		command,
		namespace,
		payload.Len(),
		corrID,
	)

	return nil
}

func (i *invoker) ExecuteMulticast(
	ctx context.Context,
	msgID overpass.MessageID,
	namespace string,
	command string,
	payload *overpass.Payload,
) error {
	msg := amqp.Publishing{
		MessageId: msgID.String(),
		Priority:  executePriority,
		Type:      command,
		Body:      payload.Bytes(),
	}
	corrID := amqputil.PutCorrelationID(ctx, &msg)

	if err := i.send(multicastExchange, namespace, msg); err != nil {
		return err
	}

	i.logger.Printf(
		"%s executed '%s' in '%s' namespace on multiple peers (%d bytes) [%s]",
		msgID.ShortString(),
		command,
		namespace,
		payload.Len(),
		corrID,
	)

	return nil
}

func (i *invoker) Done() <-chan struct{} {
	return i.done
}

func (i *invoker) Error() error {
	if err, ok := i.err.Load().(error); ok {
		return err
	}

	return nil
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
		true,  // autoAck
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
	done := i.channel.NotifyClose(make(chan *amqp.Error))

	for msg := range messages {
		i.dispatch(msg)
	}

	if amqpErr := <-done; amqpErr != nil {
		// we can't just return err when it's nil, because it will be a nil
		// *amqp.Error, as opposed to a nil "error" interface.
		i.close(amqpErr)
	} else {
		i.close(nil)
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
		message := string(msg.Body)
		if message == "" {
			message = "unknown error"
		}
		response.Error = errors.New(message)

	default:
		response.Error = fmt.Errorf(
			"malformed response, message type '%s' is unexpected",
			msg.Type,
		)
	}

	channel <- response
}

func (i *invoker) close(err error) {
	if err != nil {
		i.err.Store(err)
	}
	close(i.done)
	i.channel.Close() // TODO lock mutes?
}
