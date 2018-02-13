package options

import (
	"runtime"
	"time"

	"github.com/jmalloc/twelf/src/twelf"
	opentracing "github.com/opentracing/opentracing-go"
)

// visitor handles the application of options.
type visitor interface {
	applyDefaultTimeout(time.Duration) error
	applyLogger(twelf.Logger) error
	applyCommandWorkers(uint) error
	applySessionWorkers(uint) error
	applyPruneInterval(time.Duration) error
	applyProduct(string) error
	applyTracer(opentracing.Tracer) error
}

// Apply applies the default options, then a sequence of additional options to v.
func Apply(v visitor, opts ...Option) error {
	if err := v.applyDefaultTimeout(5 * time.Second); err != nil {
		return err
	}

	procs := uint(runtime.GOMAXPROCS(0))
	if err := v.applyCommandWorkers(procs); err != nil {
		return err
	}

	if err := v.applySessionWorkers(procs * 10); err != nil {
		return err
	}

	if err := v.applyLogger(defaultLogger); err != nil {
		return err
	}

	if err := v.applyPruneInterval(3 * time.Minute); err != nil {
		return err
	}

	if err := v.applyTracer(opentracing.NoopTracer{}); err != nil {
		return err
	}

	for _, o := range opts {
		if err := o(v); err != nil {
			return err
		}
	}

	return nil
}

var defaultLogger twelf.Logger

func init() {
	// Initialize the default logger before any testing framework can redirect
	// stdout. This lets us use standard "Output:" checks in example tests
	// without having to match the log output, while still printing the log
	// output in case of a test failure.
	defaultLogger = &twelf.StandardLogger{}
}
