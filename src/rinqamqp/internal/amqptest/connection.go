package amqptest

import (
	"os"

	"github.com/rinq/rinq-go/src/rinqamqp"
	"github.com/streadway/amqp"
)

// Connect returns a new AMQP connection for testing.
func Connect() *amqp.Connection {
	dsn := os.Getenv("RINQ_AMQP_DSN")
	if dsn == "" {
		dsn = rinqamqp.DefaultDSN
	}

	broker, err := amqp.Dial(dsn)
	if err != nil {
		panic(err)
	}

	return broker
}

// PublishingToDelivery creates a delivery message from a publishing message,
// partially simulating its transmission via the broker.
func PublishingToDelivery(msg *amqp.Publishing) *amqp.Delivery {
	return &amqp.Delivery{
		Headers:         msg.Headers,
		ContentType:     msg.ContentType,
		ContentEncoding: msg.ContentEncoding,
		DeliveryMode:    msg.DeliveryMode,
		Priority:        msg.Priority,
		CorrelationId:   msg.CorrelationId,
		ReplyTo:         msg.ReplyTo,
		Expiration:      msg.Expiration,
		MessageId:       msg.MessageId,
		Timestamp:       msg.Timestamp,
		Type:            msg.Type,
		UserId:          msg.UserId,
		AppId:           msg.AppId,
		Body:            msg.Body,
	}
}
