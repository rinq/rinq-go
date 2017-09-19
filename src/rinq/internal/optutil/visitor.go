package optutil

import (
	"runtime"
	"time"

	"github.com/rinq/rinq-go/src/rinq"
)

// Option is a function that applies a change to a visitor.
type Option func(v Visitor) error

// Visitor handles the application of options.
type Visitor interface {
	ApplyDefaultTimeout(time.Duration) error
	ApplyLogger(rinq.Logger) error
	ApplyCommandWorkers(uint) error
	ApplySessionWorkers(uint) error
	ApplyPruneInterval(time.Duration) error
	ApplyProduct(string) error
}

// Apply applies the default options, then a sequence of additional options to v.
func Apply(v Visitor, opts ...Option) error {
	if err := v.ApplyDefaultTimeout(5 * time.Second); err != nil {
		return err
	}

	procs := uint(runtime.GOMAXPROCS(0))
	if err := v.ApplyCommandWorkers(procs); err != nil {
		return err
	}

	if err := v.ApplySessionWorkers(procs * 10); err != nil {
		return err
	}

	if err := v.ApplyLogger(defaultLogger); err != nil {
		return err
	}

	if err := v.ApplyPruneInterval(3 * time.Minute); err != nil {
		return err
	}

	for _, o := range opts {
		if err := o(v); err != nil {
			return err
		}
	}

	return nil
}

var defaultLogger rinq.Logger

func init() {
	// Initialize the default logger before any testing framework can redirect
	// stdout. This lets us use standard "Output:" checks in example tests
	// without having to match the log output, while still printing the log
	// output in case of a test failure.
	defaultLogger = rinq.NewLogger(false)
}
