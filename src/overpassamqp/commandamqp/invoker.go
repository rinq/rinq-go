package commandamqp

import (
	"context"
	"errors"
	"fmt"
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
	sm *service.StateMachine

	peerID         overpass.PeerID
	preFetch       int
	defaultTimeout time.Duration
	queues         *queueSet
	channel        *amqp.Channel
	logger         overpass.Logger

	publishings   chan *publishing
	cancellations chan string // message ID
	deliveries    <-chan amqp.Delivery
	amqpClosed    chan *amqp.Error

	pending map[string]*publishing
}

// publishing encapsulates an AMQP message, the information required to publish it,
// and the channels used to respond to it.
type publishing struct {
	Exchange string
	Key      string
	Message  *amqp.Publishing

	Reply chan *amqp.Delivery
	Err   chan error
}

// newInvoker creates, initializes and returns a new invoker.
func newInvoker(
	peerID overpass.PeerID,
	preFetch int,
	defaultTimeout time.Duration,
	queues *queueSet,
	channel *amqp.Channel,
	logger overpass.Logger,
) (command.Invoker, error) {
	sm := service.NewStateMachine()

	i := &invoker{
		Service: sm,
		sm:      sm,

		peerID:         peerID,
		preFetch:       preFetch,
		defaultTimeout: defaultTimeout,
		queues:         queues,
		channel:        channel,
		logger:         logger,

		publishings:   make(chan *publishing),
		cancellations: make(chan string),
		amqpClosed:    make(chan *amqp.Error),

		pending: map[string]*publishing{},
	}

	if err := i.initialize(); err != nil {
		return nil, err
	}

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
	msg := &amqp.Publishing{
		MessageId:    msgID.String(),
		Priority:     executePriority,
		Type:         command,
		DeliveryMode: amqp.Persistent,
		Body:         payload.Bytes(),
	}
	corrID := amqputil.PutCorrelationID(ctx, msg)

	_, err := i.send(ctx, balancedExchange, namespace, msg)
	return corrID, err
}

func (i *invoker) ExecuteMulticast(
	ctx context.Context,
	msgID overpass.MessageID,
	namespace string,
	command string,
	payload *overpass.Payload,
) (string, error) {
	msg := &amqp.Publishing{
		MessageId: msgID.String(),
		Priority:  executePriority,
		Type:      command,
		Body:      payload.Bytes(),
	}
	corrID := amqputil.PutCorrelationID(ctx, msg)

	_, err := i.send(ctx, balancedExchange, namespace, msg)
	return corrID, err
}

// initialize prepares the AMQP channel and starts the state machine
func (i *invoker) initialize() error {
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

	go func() {
		logInvokerStart(i.logger, i.peerID, i.preFetch)
		err := i.sm.Run(i.run)
		logInvokerStop(i.logger, i.peerID, err)
	}()

	return nil
}

// run is the state entered when the service starts
func (i *invoker) run() (service.State, error) {
	for {
		select {
		case pub := <-i.publishings:
			i.publish(pub)

		case id := <-i.cancellations:
			delete(i.pending, id)

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

// graceful is the state entered after a graceful stop is requested
func (i *invoker) graceful() (service.State, error) {
	logInvokerStopping(i.logger, i.peerID, len(i.pending))

	for len(i.pending) > 0 {
		select {
		case id := <-i.cancellations:
			delete(i.pending, id)

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

// forceful is the state entered after a stop is requested
func (i *invoker) forceful() (service.State, error) {
	return nil, i.channel.Close()
}

// publish sends an AMQP message for a command request
func (i *invoker) publish(pub *publishing) {
	if pub.Exchange == balancedExchange {
		if _, err := i.queues.Get(i.channel, pub.Key); err != nil {
			pub.Err <- err
			return
		}
	}

	if err := i.channel.Publish(
		pub.Exchange,
		pub.Key,
		false, // mandatory
		false, // immediate
		*pub.Message,
	); err != nil {
		pub.Err <- err
		return
	}

	if pub.Reply == nil {
		close(pub.Err)
	} else {
		i.pending[pub.Message.MessageId] = pub
	}

}

// call prepares an AMQP message for use as a "call-type" command request, and
// sends it using send().
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

	payload, err := i.send(ctx, exchange, key, msg)
	return corrID, payload, err
}

// send queues an AMQP message for publication and waits for the reply (be it
// simple conformation of publication, or a full command response for
// "call-type" requests).
func (i *invoker) send(
	ctx context.Context,
	exchange string,
	key string,
	msg *amqp.Publishing,
) (*overpass.Payload, error) {
	pub := &publishing{
		Exchange: exchange,
		Key:      key,
		Message:  msg,
		Err:      make(chan error, 1),
	}

	if msg.ReplyTo == "Y" {
		pub.Reply = make(chan *amqp.Delivery, 1)
	}

	select {
	case i.publishings <- pub:
		// wait for response
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-i.sm.Graceful:
		return nil, context.Canceled
	case <-i.sm.Forceful:
		return nil, context.Canceled
	}

	select {
	case msg := <-pub.Reply:
		return i.unpack(msg)
	case err := <-pub.Err:
		return nil, err
	case <-ctx.Done():
		i.cancellations <- msg.MessageId
		return nil, ctx.Err()
	case <-i.sm.Forceful:
		return nil, context.Canceled
	}
}

// reply sends a command response to a waiting sender.
func (i *invoker) reply(msg *amqp.Delivery) {
	pub := i.pending[msg.RoutingKey]
	if pub != nil {
		delete(i.pending, msg.RoutingKey)
		pub.Reply <- msg
		msg.Ack(false)
	} else {
		msg.Reject(false)
	}
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
