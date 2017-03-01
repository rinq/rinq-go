package rinq

import (
	"runtime"
	"time"
)

// DefaultConfig is the default peer configuration.
var DefaultConfig Config

// Config describes peer configuration.
type Config struct {
	// DefaultTimeout specifies the maximum amount of time to wait for a call to
	// return. It is used if the context passed to Session.Call() does not
	// already have a deadline.
	DefaultTimeout time.Duration

	// Logger defines the target for all of the peer's logs.
	Logger Logger

	// CommandPreFetch is the number of incoming command REQUESTS that are
	// accepted at any given time. A new goroutine is started to service each
	// command request.
	CommandPreFetch int

	// SessionPreFetch is the number of command RESPONSES or notifications that
	// are buffered in memory at any given time.
	SessionPreFetch int

	// PruneInterval specifies how often the cache of remote session information
	// is purged of unused data.
	PruneInterval time.Duration
}

func init() {
	DefaultConfig = Config{
		DefaultTimeout:  5 * time.Second,
		CommandPreFetch: runtime.GOMAXPROCS(0),
		SessionPreFetch: runtime.GOMAXPROCS(0) * 10,
		Logger:          NewLogger(false),
		PruneInterval:   3 * time.Minute,
	}
}
