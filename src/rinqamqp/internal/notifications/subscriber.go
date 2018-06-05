package notifications

import (
	"sync"

	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/amqpx"
)

// Subscriber is an implementation of transport.Subscriber that adds and removes
// AMQP exchange->queue bindings for each namespace.
type Subscriber struct {
	peerID   ident.PeerID
	channels amqpx.ChannelPool

	m    sync.Mutex
	refs map[string]uint
}

// NewSubscriber returns a new AMQP-based notification subscriber.
func NewSubscriber(
	peerID ident.PeerID,
	channels amqpx.ChannelPool,
) *Subscriber {
	return &Subscriber{
		peerID:   peerID,
		channels: channels,
		refs:     map[string]uint{},
	}
}

// Listen begins listening for notifications in the ns namespace.
func (s *Subscriber) Listen(ns string) error {
	s.m.Lock()
	defer s.m.Unlock()

	n := s.refs[ns]
	s.refs[ns] = n + 1

	if n != 0 {
		return nil
	}

	c, err := s.channels.Get()
	if err != nil {
		return err
	}

	queue := notifyQueue(s.peerID)

	if err := c.QueueBind(
		queue,
		unicastRoutingKey(ns, s.peerID),
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
func (s *Subscriber) Unlisten(ns string) error {
	s.m.Lock()
	defer s.m.Unlock()

	n := s.refs[ns] - 1
	s.refs[ns] = n

	if n != 0 {
		return nil
	}

	c, err := s.channels.Get()
	if err != nil {
		return err
	}

	queue := notifyQueue(s.peerID)

	if err := c.QueueUnbind(
		queue,
		unicastRoutingKey(ns, s.peerID),
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
