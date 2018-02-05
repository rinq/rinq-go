package notifications

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/internal/notifications"
	"github.com/rinq/rinq-go/src/internal/x/bufferpool"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/amqputil"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/refactor/marshaling"
	"github.com/streadway/amqp"
)

// Sink is an AMQP based notification sink.
type Sink struct {
	channels amqputil.ChannelPool
	tracer   opentracing.Tracer
}

// NewSink returns a new notification sink.
func NewSink(
	channels amqputil.ChannelPool,
	tracer opentracing.Tracer,
) *Sink {
	return &Sink{channels, tracer}
}

// Send publishes a notification.
func (s *Sink) Send(_ context.Context, n *notifications.Notification) error {
	msg := &amqp.Publishing{}
	packCommon(msg, n)

	buf := bufferpool.Get()
	defer bufferpool.Put(buf) // hold until the message is published
	if err := marshaling.PackSpanContext(msg, s.tracer, n.SpanContext, buf); err != nil {
		return err
	}

	if n.IsMulticast {
		return s.multicast(msg, n)
	}

	return s.unicast(msg, n)
}

// unicast publishes a unicast notification.
func (s *Sink) unicast(msg *amqp.Publishing, n *notifications.Notification) error {
	packUnicastSpecific(msg, n)

	return s.publish(
		unicastExchange,
		unicastRoutingKey(n.Namespace, n.UnicastTarget.Peer),
		msg,
	)
}

// multicast publishes a multicast notification.
func (s *Sink) multicast(msg *amqp.Publishing, n *notifications.Notification) error {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf) // hold until the message is published

	packMulticastSpecific(msg, n, buf)

	return s.publish(
		multicastExchange,
		multicastRoutingKey(n.Namespace),
		msg,
	)
}

// publish publishes a message using a channel from the pool.
func (s *Sink) publish(exchange, key string, msg *amqp.Publishing) error {
	c, err := s.channels.Get()
	if err != nil {
		return err
	}
	defer s.channels.Put(c)

	return c.Publish(
		exchange,
		key,
		false, // mandatory
		false, // immediate
		*msg,
	)
}
