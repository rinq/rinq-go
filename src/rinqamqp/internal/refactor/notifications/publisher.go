package notifications

import (
	"bytes"

	"github.com/rinq/rinq-go/src/internal/transport"
	"github.com/rinq/rinq-go/src/internal/x/bufferpool"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/refactor/amqpx"
	"github.com/streadway/amqp"
)

// Publisher is an implementation of transport.Publisher that sends
// notifications as AMQP messages.
type Publisher struct {
	Channels amqpx.ChannelPool
	Encoder  *Encoder
}

// Publish sends a notification.
func (p *Publisher) Publish(n *transport.Notification) error {
	var sb, cb *bytes.Buffer

	sb = bufferpool.Get()
	defer bufferpool.Put(sb)

	if n.IsMulticast {
		cb = bufferpool.Get()
		defer bufferpool.Put(cb)
	}

	msg, err := p.Encoder.Marshal(n, sb, cb)
	if err != nil {
		return err
	}

	if n.IsMulticast {
		return p.multicast(msg, n)
	}

	return p.unicast(msg, n)
}

// unicast publishes a unicast notification.
func (p *Publisher) unicast(msg *amqp.Publishing, n *transport.Notification) error {
	return p.publish(
		unicastExchange,
		unicastRoutingKey(n.Namespace, n.UnicastTarget.Peer),
		msg,
	)
}

// multicast publishes a multicast notification.
func (p *Publisher) multicast(msg *amqp.Publishing, n *transport.Notification) error {
	return p.publish(
		multicastExchange,
		multicastRoutingKey(n.Namespace),
		msg,
	)
}

// publish publishes a message using a channel from the pool.
func (p *Publisher) publish(exchange, key string, msg *amqp.Publishing) error {
	c, err := p.Channels.Get()
	if err != nil {
		return err
	}
	defer p.Channels.Put(c)

	return c.Publish(
		exchange,
		key,
		false, // mandatory
		false, // immediate
		*msg,
	)
}
