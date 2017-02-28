package commandamqp

import (
	"context"
	"sync"
	"time"

	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpass/amqp/internal/amqputil"
	"github.com/over-pass/overpass-go/src/overpass/internal/command"
	"github.com/over-pass/overpass-go/src/overpass/internal/service"
	"github.com/over-pass/overpass-go/src/overpass/internal/trace"
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

	mutex    sync.RWMutex
	handlers map[overpass.SessionID]overpass.AsyncHandler

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

		handlers: map[overpass.SessionID]overpass.AsyncHandler{},

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
	ns string,
	cmd string,
	out *overpass.Payload,
) (string, *overpass.Payload, error) {
	msg := &amqp.Publishing{
		MessageId: msgID.String(),
		Priority:  callUnicastPriority,
	}
	packRequest(msg, ns, cmd, out, replyCorrelated)
	traceID := amqputil.PackTrace(ctx, msg)

	logUnicastCallBegin(i.logger, i.peerID, msgID, target, ns, cmd, traceID, out)
	in, err := i.call(ctx, unicastExchange, target.String(), msg)
	logCallEnd(i.logger, i.peerID, msgID, ns, cmd, traceID, in, err)

	return traceID, in, err
}

func (i *invoker) CallBalanced(
	ctx context.Context,
	msgID overpass.MessageID,
	ns string,
	cmd string,
	out *overpass.Payload,
) (string, *overpass.Payload, error) {
	msg := &amqp.Publishing{
		MessageId: msgID.String(),
		Priority:  callBalancedPriority,
	}
	packRequest(msg, ns, cmd, out, replyCorrelated)
	traceID := amqputil.PackTrace(ctx, msg)

	logBalancedCallBegin(i.logger, i.peerID, msgID, ns, cmd, traceID, out)
	in, err := i.call(ctx, balancedExchange, ns, msg)
	logCallEnd(i.logger, i.peerID, msgID, ns, cmd, traceID, in, err)

	return traceID, in, err
}

// CallBalancedAsync sends a load-balanced command request to the first
// available peer, instructs it to send a response, but does not block.
func (i *invoker) CallBalancedAsync(
	ctx context.Context,
	msgID overpass.MessageID,
	ns string,
	cmd string,
	out *overpass.Payload,
) (string, error) {
	msg := &amqp.Publishing{
		MessageId: msgID.String(),
		Priority:  callBalancedPriority,
	}
	packRequest(msg, ns, cmd, out, replyUncorrelated)
	traceID := amqputil.PackTrace(ctx, msg)

	err := i.send(ctx, balancedExchange, ns, msg)
	logAsyncRequest(i.logger, i.peerID, msgID, ns, cmd, traceID, out, err)

	return traceID, err
}

// SetAsyncHandler sets the asynchronous handler to use for a specific
// session.
func (i *invoker) SetAsyncHandler(sessID overpass.SessionID, h overpass.AsyncHandler) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if h == nil {
		delete(i.handlers, sessID)
	} else {
		i.handlers[sessID] = h
	}
}

func (i *invoker) ExecuteBalanced(
	ctx context.Context,
	msgID overpass.MessageID,
	ns string,
	cmd string,
	out *overpass.Payload,
) (string, error) {
	msg := &amqp.Publishing{
		MessageId:    msgID.String(),
		Priority:     executePriority,
		DeliveryMode: amqp.Persistent,
	}
	packRequest(msg, ns, cmd, out, replyNone)
	traceID := amqputil.PackTrace(ctx, msg)

	err := i.send(ctx, balancedExchange, ns, msg)
	logBalancedExecute(i.logger, i.peerID, msgID, ns, cmd, traceID, out, err)

	return traceID, err
}

func (i *invoker) ExecuteMulticast(
	ctx context.Context,
	msgID overpass.MessageID,
	ns string,
	cmd string,
	out *overpass.Payload,
) (string, error) {
	msg := &amqp.Publishing{
		MessageId: msgID.String(),
		Priority:  executePriority,
	}
	packRequest(msg, ns, cmd, out, replyNone)
	traceID := amqputil.PackTrace(ctx, msg)

	err := i.send(ctx, multicastExchange, ns, msg)
	logMulticastExecute(i.logger, i.peerID, msgID, ns, cmd, traceID, out, err)

	return traceID, err
}

// initialize prepares the AMQP channel and starts the state machine
func (i *invoker) initialize() error {
	if channel, err := i.channels.GetQOS(i.preFetch); err == nil { // do not return to pool, used for consume
		i.channel = channel
	} else {
		return err
	}

	i.channel.NotifyClose(i.amqpClosed)

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
	*overpass.Payload,
	error,
) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, i.defaultTimeout)
		defer cancel()
	}

	if _, err := amqputil.PackDeadline(ctx, msg); err != nil {
		return nil, err
	}

	c := call{
		msg.MessageId,
		make(chan *amqp.Delivery, 1),
	}

	select {
	case i.track <- c:
		// ready to publish
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-i.sm.Graceful:
		return nil, context.Canceled
	case <-i.sm.Forceful:
		return nil, context.Canceled
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
		return nil, err
	}

	select {
	case msg := <-c.Reply:
		payload, err := unpackResponse(msg)
		return payload, err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-i.sm.Forceful:
		return nil, context.Canceled
	}
}

// send publishes a message for a command request
func (i *invoker) send(
	ctx context.Context,
	exchange string,
	key string,
	msg *amqp.Publishing,
) error {
	select {
	default:
		return i.publish(exchange, key, msg)
	case <-ctx.Done():
		return ctx.Err()
	case <-i.sm.Graceful:
		return context.Canceled
	case <-i.sm.Forceful:
		return context.Canceled
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
	var ack bool
	if unpackReplyMode(msg) == replyUncorrelated {
		ack = i.replyAsync(msg)
	} else {
		ack = i.replySync(msg)
	}

	if ack {
		msg.Ack(false)
	} else {
		msg.Reject(false)
	}
}

func (i *invoker) replySync(msg *amqp.Delivery) bool {
	channel := i.pending[msg.RoutingKey]
	if channel == nil {
		return false
	}

	delete(i.pending, msg.RoutingKey)
	channel <- msg // buffered chan
	close(channel)

	return true
}

func (i *invoker) replyAsync(msg *amqp.Delivery) bool {
	msgID, err := overpass.ParseMessageID(msg.RoutingKey)
	if err != nil {
		logInvokerInvalidMessageID(i.logger, i.peerID, msg.RoutingKey)
		return false
	}

	ns, cmd, err := unpackNamespaceAndCommand(msg)
	if err != nil {
		logInvokerIgnoredMessage(i.logger, i.peerID, msgID, err)
		return false
	}

	i.mutex.RLock()
	handler := i.handlers[msgID.Session.ID]
	i.mutex.RUnlock()

	if handler == nil {
		return false
	}

	ctx := amqputil.UnpackTrace(context.Background(), msg)
	payload, err := unpackResponse(msg)

	logAsyncResponse(i.logger, i.peerID, msgID, ns, cmd, trace.Get(ctx), payload, err)

	go handler(ctx, msgID, ns, cmd, payload, err)

	return true
}
