package overpass

import (
	"runtime"
	"time"
)

// Config describes peer configuration.
type Config struct {
	DefaultTimeout time.Duration
	PreFetch       int
	Logger         Logger
}

// WithDefaults returns a copy of this config with empty properties replaced
// with their defaults.
func (config Config) WithDefaults() Config {
	if config.DefaultTimeout == 0 {
		config.DefaultTimeout = 5 * time.Second
	}

	if config.PreFetch == 0 {
		config.PreFetch = runtime.GOMAXPROCS(0)
	}

	if config.Logger == nil {
		config.Logger = NewLogger(false)
	}

	return config
}
