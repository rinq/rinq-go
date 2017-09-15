package rinq

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
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

// NewConfigFromEnv returns a peer configuration with values read from environment variables.
//
// The environment variables are listed below. If any variable is undefined, the
// value from DefaultConfig is used.
//
// - RINQ_DEFAULT_TIMEOUT -> Config.DefaultTimeout - in milliseconds, must be a postive integer.
// - RINQ_LOG_DEBUG       -> Enable debug logging  - must be "true" or "false".
// - RINQ_COMMAND_WORKERS -> Config.CommandWorkers - must be a positive integer.
// - RINQ_SESSION_WORKERS -> Config.SessionWorkers - must be a positive integer.
// - RINQ_PRUNE_INTERVAL  -> Config.PruneInterval  - in milliseconds, must be a postive integer.
// - RINQ_PRODUCT         -> Config.Product        - free-form string.
func NewConfigFromEnv() (Config, error) {
	cfg := DefaultConfig

	t, ok, err := timeFromEnv("RINQ_DEFAULT_TIMEOUT")
	if err != nil {
		return cfg, err
	} else if ok {
		cfg.DefaultTimeout = t
	}

	if v := os.Getenv("RINQ_LOG_DEBUG"); v != "" {
		if v == "true" {
			cfg.Logger = NewLogger(true)
		} else if v == "false" {
			cfg.Logger = NewLogger(false)
		} else {
			return cfg, errors.New("RINQ_LOG_DEBUG must be 'true' or 'false'")
		}
	}

	n, ok, err := intFromEnv("RINQ_COMMAND_WORKERS")
	if err != nil {
		return cfg, err
	} else if ok {
		cfg.CommandWorkers = n
	}

	n, ok, err = intFromEnv("RINQ_SESSION_WORKERS")
	if err != nil {
		return cfg, err
	} else if ok {
		cfg.SessionWorkers = n
	}

	t, ok, err = timeFromEnv("RINQ_PRUNE_INTERVAL")
	if err != nil {
		return cfg, err
	} else if ok {
		cfg.PruneInterval = t
	}

	cfg.Product = os.Getenv("RINQ_PRODUCT")

	return cfg, nil
}

func intFromEnv(v string) (int, bool, error) {
	if s := os.Getenv(v); s != "" {
		n, err := strconv.ParseUint(s, 10, 31)
		if err != nil {
			return 0, false, fmt.Errorf("%s must be a non-zero integer", v)
		}

		return int(n), true, nil
	}

	return 0, false, nil
}

func timeFromEnv(v string) (time.Duration, bool, error) {
	if s := os.Getenv(v); s != "" {
		n, err := strconv.ParseUint(s, 10, 63)
		if err != nil {
			return 0, false, fmt.Errorf("%s must be a non-zero duration (in milliseconds)", v)
		}

		return time.Duration(n) * time.Millisecond, true, nil
	}

	return 0, false, nil
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
