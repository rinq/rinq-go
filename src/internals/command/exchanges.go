package command

import "github.com/streadway/amqp"

const (
	// exchanceUnicast is the exchange used to publish internal command requests
	// directly to a specific peer.
	unicastExchange = "cmd.uc"

	// multicastExchange is the exchange used to publish comman requests to the
	// all peers that can service the namespace.
	multicastExchange = "cmd.mc"

	// balancedExchange is the exchange used publish command requests to the
	// first available peer that can service the namespace.
	balancedExchange = "cmd.bal"

	// responseExchange is the exchange used to publish command responses.
	responseExchange = "cmd.rsp"
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

	if err := channel.ExchangeDeclare(
		balancedExchange,
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
		responseExchange,
		"topic",
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
