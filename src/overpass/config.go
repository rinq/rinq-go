package overpass

import (
	"runtime"
	"time"
)

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

// WithDefaults returns a copy of this config with empty properties replaced
// with their defaults.
func (config Config) WithDefaults() Config {
	if config.DefaultTimeout == 0 {
		config.DefaultTimeout = 5 * time.Second
	}

	if config.CommandPreFetch == 0 {
		config.CommandPreFetch = runtime.GOMAXPROCS(0)
	}

	if config.SessionPreFetch == 0 {
		config.SessionPreFetch = runtime.GOMAXPROCS(0) * 10
	}

	if config.Logger == nil {
		config.Logger = NewLogger(false)
	}

	if config.PruneInterval == 0 {
		config.PruneInterval = 3 * time.Minute
	}

	return config
}
