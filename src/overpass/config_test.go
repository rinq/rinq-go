package overpass_test

import (
	"log"
	"os"
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
				DefaultTimeout: 5 * time.Second,
				PreFetch:       runtime.GOMAXPROCS(0),
				Logger:         log.New(os.Stdout, "", log.LstdFlags),
			}))
		})

		It("does not replace existing values", func() {
			logger := log.New(os.Stdout, "<prefix>", 0)

			config := overpass.Config{
				DefaultTimeout: 10 * time.Second,
				PreFetch:       10,
				Logger:         logger,
			}

			Expect(config.WithDefaults()).To(Equal(config))
		})
	})
})
