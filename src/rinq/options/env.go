package options

import (
	"os"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/internal/env"
)

// FromEnv returns peer options with values read from environment variables.
//
// The environment variables are listed below.
//
// - RINQ_DEFAULT_TIMEOUT (duration in milliseconds, non-zero)
// - RINQ_LOG_DEBUG       (boolean 'true' or 'false')
// - RINQ_COMMAND_WORKERS (positive integer, non-zero)
// - RINQ_SESSION_WORKERS (positive integer, non-zero)
// - RINQ_PRUNE_INTERVAL  (duration in milliseconds, non-zero)
// - RINQ_PRODUCT         (string)
func FromEnv() ([]Option, error) {
	var o []Option

	t, ok, err := env.Duration("RINQ_DEFAULT_TIMEOUT")
	if err != nil {
		return nil, err
	} else if ok {
		o = append(o, DefaultTimeout(t))
	}

	debug, ok, err := env.Bool("RINQ_LOG_DEBUG")
	if err != nil {
		return nil, err
	} else if ok {
		l := rinq.NewLogger(debug)
		o = append(o, Logger(l))
	}

	n, ok, err := env.UInt("RINQ_COMMAND_WORKERS")
	if err != nil {
		return nil, err
	} else if ok {
		o = append(o, CommandWorkers(n))
	}

	n, ok, err = env.UInt("RINQ_SESSION_WORKERS")
	if err != nil {
		return nil, err
	} else if ok {
		o = append(o, SessionWorkers(n))
	}

	t, ok, err = env.Duration("RINQ_PRUNE_INTERVAL")
	if err != nil {
		return nil, err
	} else if ok {
		o = append(o, PruneInterval(t))
	}

	if p := os.Getenv("RINQ_PRODUCT"); p != "" {
		o = append(o, Product(p))
	}

	return o, nil
}
