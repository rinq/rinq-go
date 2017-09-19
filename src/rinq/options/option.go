package options

import (
	"time"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/internal/optutil"
)

// Option is a function that applies a configuration change.
type Option = optutil.Option

// DefaultTimeout specifies the maximum amount of time to wait for a call to
// return. It is used if the context passed to Session.Call() does not
// already have a deadline.
func DefaultTimeout(t time.Duration) Option {
	return func(v optutil.Visitor) error {
		return v.ApplyDefaultTimeout(t)
	}
}

// Logger defines the target for all of the peer's logs.
func Logger(l rinq.Logger) Option {
	return func(v optutil.Visitor) error {
		return v.ApplyLogger(l)
	}
}

// CommandWorkers is the number of incoming command REQUESTS that are
// accepted at any given time. A new goroutine is started to service each
// command request.
func CommandWorkers(n uint) Option {
	return func(v optutil.Visitor) error {
		return v.ApplyCommandWorkers(n)
	}
}

// SessionWorkers is the number of command RESPONSES or notifications that
// are buffered in memory at any given time.
func SessionWorkers(n uint) Option {
	return func(v optutil.Visitor) error {
		return v.ApplySessionWorkers(n)
	}
}

// PruneInterval specifies how often the cache of remote session information
// is purged of unused data.
func PruneInterval(t time.Duration) Option {
	return func(v optutil.Visitor) error {
		return v.ApplyPruneInterval(t)
	}
}

// Product is an application-defined string that identifies the application.
// It is recommended that the product take the form "<product>/<version>"
// such as "my-app/1.3.0".
func Product(p string) Option {
	return func(v optutil.Visitor) error {
		return v.ApplyProduct(p)
	}
}