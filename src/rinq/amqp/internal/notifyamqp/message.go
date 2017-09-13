package notifyamqp

import (
	"fmt"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/streadway/amqp"
)

func packConstraint(msg *amqp.Publishing, con rinq.Constraint) {
	if msg.Headers == nil {
		msg.Headers = amqp.Table{}
	}

	for key, value := range con {
		msg.Headers[key] = value
	}
}

func unpackConstraint(msg *amqp.Delivery) (rinq.Constraint, error) {
	constraint := rinq.Constraint{}
	for key, value := range msg.Headers {
		if v, ok := value.(string); ok {
			constraint[key] = v
		} else {
			return nil, fmt.Errorf("constraint key %s contains non-string value", key)
		}
	}

	return constraint, nil
}
