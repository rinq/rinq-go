package notifications

import (
	"context"

	"github.com/jmalloc/twelf/src/twelf"
	"github.com/rinq/rinq-go/src/internal/transport"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/refactor/amqpx"
	"github.com/streadway/amqp"
)

// Consumer is a partial implementation of transport.Subscriber that decodes
// AMQP messages to produce notifications.
type Consumer struct {
	PeerID  ident.PeerID
	Broker  *amqp.Connection
	Decoder *Decoder
	Logger  twelf.Logger

	ctx           context.Context
	notifications chan<- *transport.Notification
	acks          chan ident.MessageID
	pending       map[ident.MessageID]*amqp.Delivery
}

// Consume begins accepting notifications and sends them to n.
func (c *Consumer) Consume(ctx context.Context, n chan<- *transport.Notification) error {
	defer close(n)

	preFetch := cap(n)
	if preFetch == 0 {
		panic("expected buffered channel")
	}

	c.ctx = ctx
	c.notifications = n
	c.acks = make(chan ident.MessageID, preFetch)
	c.pending = map[ident.MessageID]*amqp.Delivery{}

	ch, err := amqpx.ChannelWithPreFetch(c.Broker, preFetch)
	if err != nil {
		return err
	}
	defer ch.Close()

	closed := make(chan *amqp.Error)
	ch.NotifyClose(closed)

	queue := notifyQueue(c.PeerID)
	deliveries, err := ch.Consume(
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

	logConsumerStart(c.Logger, c.PeerID, preFetch)

	for err != nil {
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case msg := <-deliveries:
			err = c.dispatch(&msg)
		case id := <-c.acks:
			err = c.ack(id)
		case err = <-closed:
			return err // err != nil because channel is only closed in defer.
		}
	}

	logConsumerStop(c.Logger, c.PeerID, err)

	return err
}

// Ack acknowledges the notification with the given ID, it MUST be called
// after each notification has been handled.
func (c *Consumer) Ack(id ident.MessageID) {
	select {
	case <-c.ctx.Done():
	case c.acks <- id:
	}
}

func (c *Consumer) dispatch(msg *amqp.Delivery) error {
	n, err := c.Decoder.Unmarshal(msg)
	if err != nil {
		logIgnoredMessage(c.Logger, c.PeerID, msg, err)
		if err := msg.Reject(false); err != nil { // false = don't requeue
			return err
		}
	}

	c.pending[n.ID] = msg
	c.notifications <- n

	return nil
}

func (c *Consumer) ack(id ident.MessageID) error {
	if msg, ok := c.pending[id]; ok {
		delete(c.pending, id)
		return msg.Ack(false) // false = single message
	}

	return nil
}
