package overpass

import (
	"log"
	"os"
	"runtime"
	"time"
)

// Config describes peer configuration.
type Config struct {
	DefaultTimeout time.Duration
	PreFetch       int
	Logger         *log.Logger
	DebugLogging   bool
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
		config.Logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	return config
}
