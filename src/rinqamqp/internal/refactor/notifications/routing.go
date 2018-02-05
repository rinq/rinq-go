package notifications

import (
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/streadway/amqp"
)

const (
	// unicastExchange is the exchange used to publish notifications directly to
	// a specific session.
	unicastExchange = "ntf.uc"

	// multicastExchange is the exchange used to publish notifications that are
	// sent to multiple sessions based on a rinq.Constraint.
	multicastExchange = "ntf.mc"
)

// unicastRoutingKey returns the routing key used for unicast notifications.
func unicastRoutingKey(ns string, p ident.PeerID) string {
	return ns + "." + p.String()
}

// multicastRoutingKey returns the routing key used for multicast notifications.
func multicastRoutingKey(ns string) string {
	return ns
}

// notifyQueue returns the name of the queue used for incoming notifications.
func notifyQueue(p ident.PeerID) string {
	return p.ShortString() + ".ntf"
}

// DeclareResources declares all exchanges and queues used by the notification
// system for the given peer.
func DeclareResources(c *amqp.Channel, p ident.PeerID) error {
	if err := c.ExchangeDeclare(
		unicastExchange,
		"direct",
		false, // durable
		false, // autoDelete
		false, // internal
		false, // noWait
		nil,   // args
	); err != nil {
		return err
	}

	if err := c.ExchangeDeclare(
		multicastExchange,
		"direct",
		false, // durable
		false, // autoDelete
		false, // internal
		false, // noWait
		nil,   // args
	); err != nil {
		return err
	}

	if _, err := c.QueueDeclare(
		notifyQueue(p),
		false, // durable
		false, // autoDelete
		true,  // exclusive,
		false, // noWait
		nil,   // args
	); err != nil {
		return err
	}

	return nil
}
