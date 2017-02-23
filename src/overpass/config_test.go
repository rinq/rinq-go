package overpass_test

import (
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/over-pass/overpass-go/src/overpass"
)

var _ = Describe("Config", func() {
	Describe("WithDefaults", func() {
		It("returns a new config value with defaults set", func() {
			config := overpass.Config{}.WithDefaults()

			Expect(config).To(Equal(overpass.Config{
				DefaultTimeout:  5 * time.Second,
				CommandPreFetch: runtime.GOMAXPROCS(0),
				SessionPreFetch: runtime.GOMAXPROCS(0) * 10,
				Logger:          overpass.NewLogger(false),
				PruneInterval:   3 * time.Minute,
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

			Expect(config.WithDefaults()).To(Equal(config))
		})
	})
})
