package amqp

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/over-pass/overpass-go/src/overpass"
)

var _ = Describe("Dialer", func() {
	Describe("withDefaults", func() {
		It("returns a new config value with defaults set", func() {
			config := withDefaults(overpass.Config{})

			Expect(config).To(Equal(overpass.Config{
				DefaultTimeout:  overpass.DefaultConfig.DefaultTimeout,
				CommandPreFetch: overpass.DefaultConfig.CommandPreFetch,
				SessionPreFetch: overpass.DefaultConfig.SessionPreFetch,
				Logger:          overpass.DefaultConfig.Logger,
				PruneInterval:   overpass.DefaultConfig.PruneInterval,
			}))
		})

		It("does not replace existing values", func() {
			config := overpass.Config{
				DefaultTimeout:  10 * time.Second,
				CommandPreFetch: 10,
				SessionPreFetch: 20,
				Logger:          overpass.NewLogger(true),
				PruneInterval:   20 * time.Second,
			}

			Expect(withDefaults(config)).To(Equal(config))
		})
	})
})
