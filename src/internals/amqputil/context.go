package amqputil

import (
	"context"
	"strconv"
	"time"

	"github.com/streadway/amqp"
)

// PutCorrelationID sets msg.CorrelationId to the correlation ID in ctx, and
// returns the ID.
//
// If ctx does not have a correlation ID, the return value is msg.MessageId.
//
// Overpass uses the AMQP correlation ID to tie "root" requests (be they
// command requests or notifications) to any requests that are made in response
// to that "root" request. This is different to the popular use of the
// correlation ID field, which is often used to relate a response to a request.
func PutCorrelationID(ctx context.Context, msg *amqp.Publishing) string {
	id := GetCorrelationID(ctx)

	if id == "" || id == msg.MessageId {
		return msg.MessageId
	}

	msg.CorrelationId = id
	return id
}

// GetCorrelationID returns the correlation ID from ctx, or an empty string if
// none is present.
func GetCorrelationID(ctx context.Context) string {
	str, _ := ctx.Value(correlationIDKey).(string)
	return str
}

// WithCorrelationID creates a new context based on parent which includes the
// correlation ID from the given message as a context value.
//
// If the correlation ID is empty, the message is considered a "root" request,
// so the message ID is used as the correlation ID.
func WithCorrelationID(parent context.Context, msg amqp.Delivery) context.Context {
	id := msg.CorrelationId
	if id == "" {
		id = msg.MessageId
	}

	return context.WithValue(parent, correlationIDKey, id)
}

// PutExpiration uses deadline information from ctx to populate the expiration
// field of msg. The return value is true if a deadline is present.
//
// The context "done" error is returned if the deadline has already passed or
// the context has been cancelled.
func PutExpiration(ctx context.Context, msg *amqp.Publishing) (bool, error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		return false, nil
	}

	if msg.Headers == nil {
		msg.Headers = amqp.Table{}
	}

	// calculate the deadline and store it in a header
	deadlineNanos := deadline.UnixNano()
	deadlineMillis := deadlineNanos / int64(time.Millisecond)
	msg.Headers[deadlineHeader] = deadlineMillis

	// calculate the expiration based on current time
	msg.Expiration = "0"
	remainingMillis := deadline.Sub(time.Now()) / time.Millisecond

	select {
	case <-ctx.Done():
		return true, ctx.Err()
	default:
		if remainingMillis > 0 {
			msg.Expiration = strconv.FormatInt(int64(remainingMillis), 10)
		}
		return true, nil
	}
}

// WithExpiration creates a new context based on parent which has a deadline
// computed from the expiration information in msg.
//
// The return values are the same as context.WithDeadline()
func WithExpiration(parent context.Context, msg amqp.Delivery) (context.Context, func()) {
	deadlineMillis, ok := msg.Headers[deadlineHeader].(int64)
	if !ok {
		return context.WithCancel(parent)
	}

	deadlineNanos := deadlineMillis * int64(time.Millisecond)
	deadline := time.Unix(0, deadlineNanos)

	return context.WithDeadline(parent, deadline)
}

type contextKey string

var correlationIDKey = contextKey("correlation-id")

const deadlineHeader = "dl"
