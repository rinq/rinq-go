package trace

import "context"

// With returns a new context derived from parent that includes an
// application-defined "trace ID" value.
//
// Any operations (such as command calls, session notifications, etc) that use
// the returned context will include the trace ID in log output on both the
// sending and receiving peer.
//
// The trace ID is also present in the ctx supplied to command and notification
// handlers. This allows very easy propagation of the trace identifier to all
// "sub-requests" of the initial operation.
//
// The trace ID is rendered surrounded by square brackets in log output.
//
// If an operation is performed with ctx that does not contain a trace ID,
// the operation's message ID is used. This includes sufficient information to
// identify the peer, session and revision of the intial operation.
func With(parent context.Context, t string) context.Context {
	return context.WithValue(parent, key, t)
}

// Get returns the trace identifier from ctx, or an empty string if none is
// present.
func Get(ctx context.Context) string {
	str, _ := ctx.Value(key).(string)
	return str
}

type keyType struct{}

var key keyType
