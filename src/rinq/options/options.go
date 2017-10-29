package options

import (
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/rinq"
)

// Options is a structure representing a resolved set of options.
type Options struct {
	DefaultTimeout time.Duration
	Logger         rinq.Logger
	CommandWorkers uint
	SessionWorkers uint
	PruneInterval  time.Duration
	Product        string
	Tracer         opentracing.Tracer
}

// NewOptions returns a new Options object from the given options, with default
// values for any options that are not specified.
func NewOptions(opts ...Option) (o Options, err error) {
	err = Apply(&o, opts...)
	return
}

// applyDefaultTimeout sets the DefaultTimeout value.
func (o *Options) applyDefaultTimeout(v time.Duration) error {
	o.DefaultTimeout = v
	return nil
}

// applyLogger sets the Logger value.
func (o *Options) applyLogger(v rinq.Logger) error {
	if v == nil {
		panic("logger must not be nil")
	}

	o.Logger = v
	return nil
}

// applyCommandWorkers sets the CommandWorkers value.
func (o *Options) applyCommandWorkers(v uint) error {
	o.CommandWorkers = v
	return nil
}

// applySessionWorkers sets the SessionWorkers value.
func (o *Options) applySessionWorkers(v uint) error {
	o.SessionWorkers = v
	return nil
}

// applyPruneInterval sets the PruneInterval value.
func (o *Options) applyPruneInterval(v time.Duration) error {
	o.PruneInterval = v
	return nil
}

// applyProduct sets the Product value.
func (o *Options) applyProduct(v string) error {
	o.Product = v
	return nil
}

// applyTracer sets the Tracer value.
func (o *Options) applyTracer(v opentracing.Tracer) error {
	if v == nil {
		panic("tracer must not be nil")
	}

	o.Tracer = v
	return nil
}
