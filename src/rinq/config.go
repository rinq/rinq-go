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

	// CommandWorkers is the number of incoming command REQUESTS that are
	// accepted at any given time. A new goroutine is started to service each
	// command request.
	CommandWorkers int

	// SessionWorkers is the number of command RESPONSES or notifications that
	// are buffered in memory at any given time.
	SessionWorkers int

	// PruneInterval specifies how often the cache of remote session information
	// is purged of unused data.
	PruneInterval time.Duration

	// Product is an application-defined string that identifies the application.
	// It is recommended that the product take the form "<product>/<version>"
	// such as "my-app/1.3.0".
	Product string
}

func init() {
	DefaultConfig = Config{
		DefaultTimeout: 5 * time.Second,
		CommandWorkers: runtime.GOMAXPROCS(0),
		SessionWorkers: runtime.GOMAXPROCS(0) * 10,
		Logger:         NewLogger(false),
		PruneInterval:  3 * time.Minute,
	}
}
