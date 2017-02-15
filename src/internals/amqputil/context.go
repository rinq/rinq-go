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
// Overpass uses the AMQP correlation ID to track a "root" request (be it a
// command execution or notification) across the entire network, included any
// additional requests made in response to the "root" request. This is contrary
// to the popular use of the correlation ID field, which is often used to relate
// a response to a request.
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

	msg.Timestamp = time.Now()
	msg.Expiration = "0"

	remaining := deadline.Sub(msg.Timestamp) / time.Millisecond

	select {
	case <-ctx.Done():
		return true, ctx.Err()
	default:
		if remaining > 0 {
			msg.Expiration = strconv.FormatInt(int64(remaining), 10)
		}
		return true, nil
	}
}

// WithExpiration creates a new context based on parent which has a deadline
// computed from the expiration information in msg.
//
// The return values are the same as context.WithDeadline()
func WithExpiration(parent context.Context, msg amqp.Delivery) (context.Context, func()) {
	if msg.Timestamp.IsZero() {
		return context.WithCancel(parent)
	}

	ttl, err := strconv.ParseUint(msg.Expiration, 10, 64)
	if err != nil {
		return context.WithCancel(parent)
	}

	return context.WithDeadline(
		parent,
		msg.Timestamp.Add(time.Duration(ttl)),
	)
}

type contextKey string

var correlationIDKey = contextKey("correlation-id")
