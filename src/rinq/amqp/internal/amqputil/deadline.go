package amqputil

import (
	"context"
	"strconv"
	"time"

	"github.com/streadway/amqp"
)

// PackDeadline uses deadline information from ctx to populate the expiration
// field of msg. The return value is true if a deadline is present.
//
// The context "done" error is returned if the deadline has already passed or
// the context has been canceled.
func PackDeadline(ctx context.Context, msg *amqp.Publishing) (bool, error) {
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

// UnpackDeadline creates a new context based on parent which has a deadline
// computed from the expiration information in msg.
//
// The return values are the same as context.WithDeadline()
func UnpackDeadline(parent context.Context, msg *amqp.Delivery) (context.Context, func()) {
	deadlineMillis, ok := msg.Headers[deadlineHeader].(int64)
	if !ok {
		return context.WithCancel(parent)
	}

	deadlineNanos := deadlineMillis * int64(time.Millisecond)
	deadline := time.Unix(0, deadlineNanos)

	return context.WithDeadline(parent, deadline)
}

const deadlineHeader = "dl"
