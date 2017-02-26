package commandamqp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/over-pass/overpass-go/src/internal/amqputil"
	"github.com/over-pass/overpass-go/src/internal/command"
	"github.com/over-pass/overpass-go/src/internal/service"
	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/streadway/amqp"
)

// invoker is an AMQP-based implementation of command.Invoker
type invoker struct {
	service.Service
	sm *service.StateMachine

	peerID         overpass.PeerID
	preFetch       int
	defaultTimeout time.Duration
	queues         *queueSet
	channels       amqputil.ChannelPool
	channel        *amqp.Channel // channel used for consuming
	logger         overpass.Logger

	track      chan call            // add information about a call to pending
	cancel     chan call            // remove call information from pending
	deliveries <-chan amqp.Delivery // incoming command responses
	amqpClosed chan *amqp.Error

	// state-machine data
	pending map[string]chan *amqp.Delivery // map of message ID to reply channel
}

// call associates the message ID of a command request with the AMQP channel
// used to deliver the response.
type call struct {
	ID    string
	Reply chan *amqp.Delivery
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
	i := &invoker{
		peerID:         peerID,
		preFetch:       preFetch,
		defaultTimeout: defaultTimeout,
		queues:         queues,
		channels:       channels,
		logger:         logger,

		track:      make(chan call),
		cancel:     make(chan call),
		amqpClosed: make(chan *amqp.Error, 1),

		pending: map[string]chan *amqp.Delivery{},
	}

	i.sm = service.NewStateMachine(i.run, i.finalize)
	i.Service = i.sm

	if err := i.initialize(); err != nil {
		return nil, err
	}

	go i.sm.Run()

	return i, nil
}

func (i *invoker) CallUnicast(
	ctx context.Context,
	msgID overpass.MessageID,
	target overpass.PeerID,
	namespace string,
	command string,
	payload *overpass.Payload,
) (string, *overpass.Payload, error) {
	return i.call(ctx, unicastExchange, target.String(), &amqp.Publishing{
		MessageId: msgID.String(),
		Priority:  callUnicastPriority,
		Type:      command,
		Headers:   amqp.Table{namespaceHeader: namespace},
		Body:      payload.Bytes(),
	})
}

func (i *invoker) CallBalanced(
	ctx context.Context,
	msgID overpass.MessageID,
	namespace string,
	command string,
	payload *overpass.Payload,
) (string, *overpass.Payload, error) {
	return i.call(ctx, balancedExchange, namespace, &amqp.Publishing{
		MessageId: msgID.String(),
		Priority:  callBalancedPriority,
		Type:      command,
		Body:      payload.Bytes(),
	})
}

func (i *invoker) ExecuteBalanced(
	ctx context.Context,
	msgID overpass.MessageID,
	namespace string,
	command string,
	payload *overpass.Payload,
) (string, error) {
	return i.execute(ctx, balancedExchange, namespace, &amqp.Publishing{
		MessageId:    msgID.String(),
		Priority:     executePriority,
		Type:         command,
		DeliveryMode: amqp.Persistent,
		Body:         payload.Bytes(),
	})
}

func (i *invoker) ExecuteMulticast(
	ctx context.Context,
	msgID overpass.MessageID,
	namespace string,
	command string,
	payload *overpass.Payload,
) (string, error) {
	return i.execute(ctx, multicastExchange, namespace, &amqp.Publishing{
		MessageId: msgID.String(),
		Priority:  executePriority,
		Type:      command,
		Body:      payload.Bytes(),
	})
}

// initialize prepares the AMQP channel and starts the state machine
func (i *invoker) initialize() error {
	if channel, err := i.channels.Get(); err == nil { // do not return to pool, used for consume
		i.channel = channel
	} else {
		return err
	}

	i.channel.NotifyClose(i.amqpClosed)

	if err := i.channel.Qos(i.preFetch, 0, true); err != nil {
		return err
	}

	queue := responseQueue(i.peerID)

	if _, err := i.channel.QueueDeclare(
		queue,
		false, // durable
		false, // autoDelete
		true,  // exclusive,
		false, // noWait
		nil,   // args
	); err != nil {
		return err
	}

	if err := i.channel.QueueBind(
		queue,
		i.peerID.String()+".*",
		responseExchange,
		false, // noWait
		nil,   // args
	); err != nil {
		return err
	}

	var err error
	i.deliveries, err = i.channel.Consume(
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

	return nil
}

// run is the state entered when the service starts
func (i *invoker) run() (service.State, error) {
	logInvokerStart(i.logger, i.peerID, i.preFetch)

	for {
		select {
		case c := <-i.track:
			i.pending[c.ID] = c.Reply

		case c := <-i.cancel:
			delete(i.pending, c.ID)

		case msg, ok := <-i.deliveries:
			if !ok {
				// sometimes the consumer channel is closed before the AMQP channel
				return nil, <-i.amqpClosed
			}
			i.reply(&msg)

		case <-i.sm.Graceful:
			return i.graceful, nil

		case <-i.sm.Forceful:
			return i.forceful, nil

		case err := <-i.amqpClosed:
			return nil, err
		}
	}
}

// graceful is the state entered when a graceful stop is requested
func (i *invoker) graceful() (service.State, error) {
	logInvokerStopping(i.logger, i.peerID, len(i.pending))

	for len(i.pending) > 0 {
		select {
		case c := <-i.cancel:
			delete(i.pending, c.ID)

		case msg, ok := <-i.deliveries:
			if !ok {
				// sometimes the consumer channel is closed before the AMQP channel
				return nil, <-i.amqpClosed
			}
			i.reply(&msg)

		case <-i.sm.Forceful:
			return i.forceful, nil

		case err := <-i.amqpClosed:
			return nil, err
		}
	}

	return i.forceful, nil
}

// forceful is the state entered when a stop is requested
func (i *invoker) forceful() (service.State, error) {
	return nil, i.channel.Close()
}

// finalize is the state-machine finalizer, it is called immediately before the
// Done() channel is closed.
func (i *invoker) finalize(err error) error {
	logInvokerStop(i.logger, i.peerID, err)
	return err
}

// call publishes a message for an "call-type" invocation and awaits the response
func (i *invoker) call(
	ctx context.Context,
	exchange string,
	key string,
	msg *amqp.Publishing,
) (
	string,
	*overpass.Payload,
	error,
) {
	corrID := amqputil.PutCorrelationID(ctx, msg)

	if _, ok := ctx.Deadline(); !ok {
		var cancel func()
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()
	}

	msg.ReplyTo = "Y"
	if _, err := amqputil.PutExpiration(ctx, msg); err != nil {
		return corrID, nil, err
	}

	c := call{
		msg.MessageId,
		make(chan *amqp.Delivery, 1),
	}

	select {
	case i.track <- c:
		// ready to publish
	case <-ctx.Done():
		return corrID, nil, ctx.Err()
	case <-i.sm.Graceful:
		return corrID, nil, context.Canceled
	case <-i.sm.Forceful:
		return corrID, nil, context.Canceled
	}

	// notify the state machine that we're bailing if it hasn't already sent
	// us our reply
	defer func() {
		select {
		case <-c.Reply:
		default:
			select {
			case i.cancel <- c:
			case <-i.sm.Forceful:
			}
		}
	}()

	err := i.publish(exchange, key, msg)
	if err != nil {
		return corrID, nil, err
	}

	select {
	case msg := <-c.Reply:
		payload, err := i.unpack(msg)
		return corrID, payload, err
	case <-ctx.Done():
		return corrID, nil, ctx.Err()
	case <-i.sm.Forceful:
		return corrID, nil, context.Canceled
	}
}

// execute publishes a message for an "execute-type" invocation
func (i *invoker) execute(
	ctx context.Context,
	exchange string,
	key string,
	msg *amqp.Publishing,
) (string, error) {
	corrID := amqputil.PutCorrelationID(ctx, msg)

	select {
	default:
		return corrID, i.publish(exchange, key, msg)
	case <-ctx.Done():
		return corrID, ctx.Err()
	case <-i.sm.Graceful:
		return corrID, context.Canceled
	case <-i.sm.Forceful:
		return corrID, context.Canceled
	}
}

// publish sends an command request to the broker
func (i *invoker) publish(
	exchange string,
	key string,
	msg *amqp.Publishing,
) error {
	channel, err := i.channels.Get()
	if err != nil {
		return err
	}
	defer i.channels.Put(channel)

	if exchange == balancedExchange {
		if _, err = i.queues.Get(channel, key); err != nil {
			return err
		}
	}

	return channel.Publish(
		exchange,
		key,
		false, // mandatory
		false, // immediate
		*msg,
	)
}

// reply sends a command response to a waiting sender.
func (i *invoker) reply(msg *amqp.Delivery) {
	channel := i.pending[msg.RoutingKey]
	if channel == nil {
		msg.Reject(false)
		return
	}

	msg.Ack(false)

	delete(i.pending, msg.RoutingKey)
	channel <- msg // buffered chan
	close(channel)
}

// unpack extracts the payload and error information from an AMQP message.
func (i *invoker) unpack(msg *amqp.Delivery) (*overpass.Payload, error) {
	switch msg.Type {
	case successResponse:
		return overpass.NewPayloadFromBytes(msg.Body), nil

	case failureResponse:
		failureType, _ := msg.Headers[failureTypeHeader].(string)
		if failureType == "" {
			return nil, errors.New("malformed response, failure type must be a non-empty string")
		}

		payload := overpass.NewPayloadFromBytes(msg.Body)
		return payload, overpass.Failure{
			Type:    failureType,
			Message: msg.Headers[failureMessageHeader].(string),
			Payload: payload,
		}

	case errorResponse:
		return nil, overpass.UnexpectedError(msg.Body)

	default:
		return nil, fmt.Errorf("malformed response, message type '%s' is unexpected", msg.Type)
	}
}
