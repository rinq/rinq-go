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
// identify the peer, session and revision of the initial operation.
func With(parent context.Context, t string) context.Context {
	return context.WithValue(parent, key, t)
}

// WithRoot returns a new context derived from parent that includes an
// application-defined "trace ID" value, only if the parent does not already
// contain such a trace ID.
//
// If parent already contains a trace ID, ctx is parent, id is the trace ID from
// parent and ok is false. Otherwise, ctx is the derived context containing  t
// as the trace ID, id is t and ok is true.
func WithRoot(parent context.Context, t string) (ctx context.Context, id string, ok bool) {
	existing := parent.Value(key)

	if existing == nil {
		return With(parent, t), t, true
	}

	return parent, existing.(string), false
}

// Get returns the trace identifier from ctx, or an empty string if none is
// present.
func Get(ctx context.Context) string {
	str, _ := ctx.Value(key).(string)
	return str
}

type keyType struct{}

var key keyType
