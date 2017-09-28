package optutil

import (
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/rinq"
)

// Config is a resolved set of options.
type Config struct {
	DefaultTimeout time.Duration
	Logger         rinq.Logger
	CommandWorkers uint
	SessionWorkers uint
	PruneInterval  time.Duration
	Product        string
	Tracer         opentracing.Tracer
}

// NewConfig returns a new config object from the given options.
func NewConfig(opts ...Option) (cfg Config, err error) {
	err = Apply(&cfg, opts...)
	return
}

// ApplyDefaultTimeout sets the DefaultTimeout value.
func (c *Config) ApplyDefaultTimeout(v time.Duration) error {
	c.DefaultTimeout = v
	return nil
}

// ApplyLogger sets the Logger value.
func (c *Config) ApplyLogger(v rinq.Logger) error {
	if v == nil {
		panic("logger must not be nil")
	}

	c.Logger = v
	return nil
}

// ApplyCommandWorkers sets the CommandWorkers value.
func (c *Config) ApplyCommandWorkers(v uint) error {
	c.CommandWorkers = v
	return nil
}

// ApplySessionWorkers sets the SessionWorkers value.
func (c *Config) ApplySessionWorkers(v uint) error {
	c.SessionWorkers = v
	return nil
}

// ApplyPruneInterval sets the PruneInterval value.
func (c *Config) ApplyPruneInterval(v time.Duration) error {
	c.PruneInterval = v
	return nil
}

// ApplyProduct sets the Product value.
func (c *Config) ApplyProduct(v string) error {
	c.Product = v
	return nil
}

// ApplyTracer sets the Tracer value.
func (c *Config) ApplyTracer(v opentracing.Tracer) error {
	if v == nil {
		panic("tracer must not be nil")
	}

	c.Tracer = v
	return nil
}
