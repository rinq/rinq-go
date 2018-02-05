package notifications

import (
	"context"
	"fmt"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/internal/notifications"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/refactor/amqpx"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/refactor/marshaling"
	"github.com/streadway/amqp"
)

// Source is an interface that produces notifications.
type Source struct {
	peerID  ident.PeerID
	channel *amqp.Channel
	logger  rinq.Logger
	tracer  opentracing.Tracer

	amqpError  chan *amqp.Error
	deliveries <-chan amqp.Delivery

	namespaces       map[string]int
	listen, unlisten chan string
	notifications    chan *notifications.Inbound
	done             chan struct{}
}

// NewSource returns a new notification source.
func NewSource(
	peerID ident.PeerID,
	broker *amqp.Connection,
	preFetch uint,
	logger rinq.Logger,
	tracing opentracing.Tracer,
) (*Source, error) {
	c, err := amqpx.ChannelWithQOS(broker, preFetch)
	if err != nil {
		return nil, err
	}

	return &Source{
		peerID:  peerID,
		channel: c,
		logger:  logger,
		tracer:  tracing,

		namespaces:    map[string]int{},
		listen:        make(chan string),
		unlisten:      make(chan string),
		notifications: make(chan *notifications.Inbound, preFetch),
		done:          make(chan struct{}),
	}, nil
}

// Listen begins listening for notifications in the ns namespace.
func (s *Source) Listen(ns string) {
	select {
	case s.listen <- ns:
	case <-s.done:
	}
}

// Unlisten stops listening for notifications in the ns namespace.
func (s *Source) Unlisten(ns string) {
	select {
	case s.unlisten <- ns:
	case <-s.done:
	}
}

// Notifications returns a channel on which incoming notifications are delivered.
func (s *Source) Notifications() <-chan *notifications.Inbound {
	return s.notifications
}

// Run listens for messages until ctx is done.
func (s *Source) Run(ctx context.Context) error {
	defer close(s.notifications)
	defer close(s.done)

	err := s.startConsumer()

	if err == nil {
		logSourceStart(s.logger, s.peerID, cap(s.notifications))

		for err == nil {
			select {
			case ns := <-s.listen:
				err = s.bind(ns)
			case ns := <-s.unlisten:
				err = s.unbind(ns)
			case msg, ok := <-s.deliveries:
				if ok {
					err = s.emit(ctx, &msg)
				} else {
					// if the consumer channel is closed before we shutdown, there
					// must be an AMQP error.
					err = <-s.amqpError
				}
			case err = <-s.amqpError:
			case <-ctx.Done():
				err = ctx.Err()
			}
		}

		// canceling the context is the standard way to stop the source
		// and does not indicate an error.
		if err == context.Canceled {
			err = nil
		}
	}

	if e := s.stopConsumer(); e != nil {
		if err == nil {
			// only report e if there's no causal error.
			err = e
		}
	}

	logSourceStop(s.logger, s.peerID, err)

	return err
}

// emit sends a notification to the s.Notifications() channel, based on msg.
func (s *Source) emit(ctx context.Context, msg *amqp.Delivery) error {
	n := &notifications.Inbound{
		Ack: func() {
			_ = msg.Ack(false) // false = single message
		},
	}

	if err := s.unpack(msg, n); err != nil {
		logIgnoredMessage(s.logger, s.peerID, msg, err)
		_ = msg.Reject(false) // false = don't requeue
		return nil
	}

	select {
	case s.notifications <- n:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// unpack unpacks msg into n.
func (s *Source) unpack(msg *amqp.Delivery, n *notifications.Inbound) error {
	if err := unpackCommon(msg, &n.Common); err != nil {
		return err
	}

	var err error
	n.SpanContext, err = marshaling.UnpackSpanContext(msg, s.tracer)
	if err != nil {
		logIgnoredSpanContext(s.logger, s.peerID, msg, err)
	}

	switch msg.Exchange {
	case unicastExchange:
		if err := unpackUnicastSpecific(msg, &n.Common); err != nil {
			return err
		}

	case multicastExchange:
		n.IsMulticast = true
		if err := unpackMulticastSpecific(msg, &n.Common); err != nil {
			return err
		}

	default:
		return fmt.Errorf("delivery via '%s' exchange is not expected", msg.Exchange)
	}

	return nil
}

// startConsumer starts the AMQP consumer that receives notification messages.
func (s *Source) startConsumer() error {
	s.amqpError = make(chan *amqp.Error, 1)
	s.channel.NotifyClose(s.amqpError)

	var err error
	queue := notifyQueue(s.peerID)

	s.deliveries, err = s.channel.Consume(
		queue,
		queue, // use queue name as consumer tag
		false, // autoAck
		true,  // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	return err
}

// stopConsumer cancels the AMQP consumer and rejects any new messages that are
// received while the cancelation is in progress.
func (s *Source) stopConsumer() error {
	if err := s.channel.Cancel(
		notifyQueue(s.peerID), // use queue name as consumer tag
		false, // noWait
	); err != nil {
		return err
	}

	// reject any messages that have already been delivered.
	// s.deliveries is eventually closed due to the call to Cancel() above.
	for msg := range s.deliveries {
		if err := msg.Reject(false); err != nil { // false = don't requeue
			return err
		}
	}

	return nil
}

// bind adds a binding from the unicast exchange to this peer's notification
// queue.
func (s *Source) bind(ns string) error {
	count := s.namespaces[ns]
	s.namespaces[ns] = count + 1

	if count != 0 {
		return nil
	}

	queue := notifyQueue(s.peerID)

	if err := s.channel.QueueBind(
		queue,
		unicastRoutingKey(ns, s.peerID),
		unicastExchange,
		false, // noWait
		nil,   // args
	); err != nil {
		return err
	}

	return s.channel.QueueBind(
		queue,
		multicastRoutingKey(ns),
		multicastExchange,
		false, // noWait
		nil,   // args
	)
}

// unbind removes a binding from the unicast exchange to this peer's
// notification queue. unbind() must be called once for each prior call to
// bind() before the AMQP binding is actually removed.
func (s *Source) unbind(ns string) error {
	count := s.namespaces[ns] - 1
	s.namespaces[ns] = count

	if count != 0 {
		return nil
	}

	queue := notifyQueue(s.peerID)

	if err := s.channel.QueueUnbind(
		queue,
		unicastRoutingKey(ns, s.peerID),
		unicastExchange,
		nil, // args
	); err != nil {
		return err
	}

	return s.channel.QueueUnbind(
		queue,
		multicastRoutingKey(ns),
		multicastExchange,
		nil, // args
	)
}
