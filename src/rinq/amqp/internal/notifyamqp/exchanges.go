package notifyamqp

import "github.com/streadway/amqp"

const (
	// unicastExchange is the exchange used to publish notifications directly to
	// a specific session.
	unicastExchange = "ntf.uc"

	// multicastExchange is the exchange used to publish notifications that are
	// sent to target multiple sessions based on an rinq.Constraint.
	multicastExchange = "ntf.mc"
)

func declareExchanges(channel *amqp.Channel) error {
	if err := channel.ExchangeDeclare(
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

	if err := channel.ExchangeDeclare(
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

	return nil
}
