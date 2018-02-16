package notifications

import (
	"sync"

	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/refactor/amqpx"
)

// Binder is a partial implementation of transport.Subscriber that adds and
// removes AMQP exchange->queue bindings for each namespace.
type Binder struct {
	PeerID   ident.PeerID
	Channels amqpx.ChannelPool

	m    sync.Mutex
	refs map[string]uint
}

// Listen begins listening for notifications in the ns namespace.
func (b *Binder) Listen(ns string) error {
	b.m.Lock()
	defer b.m.Unlock()

	n := b.refs[ns]
	b.refs[ns] = n + 1

	if n != 0 {
		return nil
	}

	c, err := b.Channels.Get()
	if err != nil {
		return err
	}

	queue := notifyQueue(b.PeerID)

	if err := c.QueueBind(
		queue,
		unicastRoutingKey(ns, b.PeerID),
		unicastExchange,
		false, // noWait
		nil,   // args
	); err != nil {
		return err
	}

	return c.QueueBind(
		queue,
		multicastRoutingKey(ns),
		multicastExchange,
		false, // noWait
		nil,   // args
	)
}

// Unlisten stops listening for notifications in the ns namespace.
// It must be called once for each prior call to Listen() before
// notifications from ns are stopped.
func (b *Binder) Unlisten(ns string) error {
	b.m.Lock()
	defer b.m.Unlock()

	n := b.refs[ns] - 1
	b.refs[ns] = n

	if n != 0 {
		return nil
	}

	c, err := b.Channels.Get()
	if err != nil {
		return err
	}

	queue := notifyQueue(b.PeerID)

	if err := c.QueueUnbind(
		queue,
		unicastRoutingKey(ns, b.PeerID),
		unicastExchange,
		nil, // args
	); err != nil {
		return err
	}

	return c.QueueUnbind(
		queue,
		multicastRoutingKey(ns),
		multicastExchange,
		nil, // args
	)
}
