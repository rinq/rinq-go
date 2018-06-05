package marshaling

import (
	"github.com/streadway/amqp"
)

// PackTrace packs the "trace ID" into msg.
//
// As a minor optimisation, if the trace ID is the same as the message ID, it
// is not sent across the wire.
func PackTrace(msg *amqp.Publishing, traceID string) {
	if traceID != msg.MessageId {
		msg.CorrelationId = traceID
	}
}

// UnpackTrace returns the trace ID from msg.
//
// If the correlation ID is empty, the message is considered a "root" request,
// so the message ID is used as the correlation ID.
func UnpackTrace(msg *amqp.Delivery) string {
	if msg.CorrelationId != "" {
		return msg.CorrelationId
	}

	return msg.MessageId
}
