package amqputil

import (
	"context"

	"github.com/over-pass/overpass-go/src/overpass/internal/trace"
	"github.com/streadway/amqp"
)

// PackTrace sets msg.CorrelationId to the trace ID in ctx, and returns the ID.
//
// If ctx does not have a trace ID, the return value is msg.MessageId.
//
// Overpass uses the AMQP correlation ID to tie "root" requests (be they
// command requests or notifications) to any requests that are made in response
// to that "root" request. This is different to the popular use of the
// correlation ID field, which is often used to relate a response to a request.
func PackTrace(ctx context.Context, msg *amqp.Publishing) string {
	traceID := trace.Get(ctx)

	if traceID == "" || traceID == msg.MessageId {
		return msg.MessageId
	}

	msg.CorrelationId = traceID
	return traceID
}

// UnpackTrace creates a new context with a trace ID based on the AMQP correlation
// ID from msg.
//
// If the correlation ID is empty, the message is considered a "root" request,
// so the message ID is used as the correlation ID.
func UnpackTrace(parent context.Context, msg *amqp.Delivery) context.Context {
	if msg.CorrelationId != "" {
		return trace.With(parent, msg.CorrelationId)
	}

	return trace.With(parent, msg.MessageId)
}
