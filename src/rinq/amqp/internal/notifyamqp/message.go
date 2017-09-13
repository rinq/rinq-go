package notifyamqp

import (
	"errors"
	"fmt"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/streadway/amqp"
)

const (
	// constraintHeader specifies the constraint for multicast notifications.
	constraintHeader = "c"
)

func packConstraint(msg *amqp.Publishing, con rinq.Constraint) {
	t := amqp.Table{}
	for key, value := range con {
		t[key] = value
	}

	if msg.Headers == nil {
		msg.Headers = amqp.Table{}
	}

	msg.Headers[constraintHeader] = t
}

func unpackConstraint(msg *amqp.Delivery) (rinq.Constraint, error) {
	if t, ok := msg.Headers[constraintHeader].(amqp.Table); ok {
		con := rinq.Constraint{}

		for key, value := range t {
			if v, ok := value.(string); ok {
				con[key] = v
			} else {
				return nil, fmt.Errorf("constraint key %s contains non-string value", key)
			}
		}

		return con, nil
	}

	return nil, errors.New("constraint header is not a table")
}
