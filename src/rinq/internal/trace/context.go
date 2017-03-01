package trace

import "context"

// Get returns the trace identifier from ctx, or an empty string if none is
// present.
func Get(ctx context.Context) string {
	str, _ := ctx.Value(key).(string)
	return str
}

// With creates a new context based on parent with t as the trace identifier.
func With(parent context.Context, t string) context.Context {
	return context.WithValue(parent, key, t)
}

type keyType struct{}

var key keyType
