package rinq

import (
	"os"
	"runtime"
	"time"

	"github.com/rinq/rinq-go/src/rinq/internal/env"
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

// NewConfigFromEnv returns a peer configuration with values read from environment variables.
//
// The environment variables are listed below. Each variable directly affects
// one of the fields in the Config struct. If any variable is undefined, the
// value from DefaultConfig is used.
//
// - RINQ_DEFAULT_TIMEOUT (duration in milliseconds, non-zero)
// - RINQ_LOG_DEBUG       (boolean 'true' or 'false')
// - RINQ_COMMAND_WORKERS (positive integer, non-zero)
// - RINQ_SESSION_WORKERS (positive integer, non-zero)
// - RINQ_PRUNE_INTERVAL  (duration in milliseconds, non-zero)
// - RINQ_PRODUCT         (string)
func NewConfigFromEnv() (cfg Config, err error) {
	cfg.DefaultTimeout, err = env.Duration("RINQ_DEFAULT_TIMEOUT", DefaultConfig.DefaultTimeout)
	if err != nil {
		return
	}

	debug, err := env.Bool("RINQ_LOG_DEBUG", false)
	if err != nil {
		return
	}

	if debug == DefaultConfig.Logger.IsDebug() {
		cfg.Logger = DefaultConfig.Logger
	} else {
		cfg.Logger = NewLogger(debug)
	}

	cfg.CommandWorkers, err = env.Int("RINQ_COMMAND_WORKERS", DefaultConfig.CommandWorkers)
	if err != nil {
		return
	}

	cfg.SessionWorkers, err = env.Int("RINQ_SESSION_WORKERS", DefaultConfig.SessionWorkers)
	if err != nil {
		return
	}

	cfg.PruneInterval, err = env.Duration("RINQ_PRUNE_INTERVAL", DefaultConfig.PruneInterval)
	if err != nil {
		return
	}

	cfg.Product = os.Getenv("RINQ_PRODUCT")

	return
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
