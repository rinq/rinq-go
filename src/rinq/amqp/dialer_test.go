package amqp

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("Dialer", func() {
	Describe("withDefaults", func() {
		It("returns a new config value with defaults set", func() {
			config := withDefaults(rinq.Config{})

			Expect(config).To(Equal(rinq.Config{
				DefaultTimeout:  rinq.DefaultConfig.DefaultTimeout,
				CommandPreFetch: rinq.DefaultConfig.CommandPreFetch,
				SessionPreFetch: rinq.DefaultConfig.SessionPreFetch,
				Logger:          rinq.DefaultConfig.Logger,
				PruneInterval:   rinq.DefaultConfig.PruneInterval,
			}))
		})

		It("does not replace existing values", func() {
			config := rinq.Config{
				DefaultTimeout:  10 * time.Second,
				CommandPreFetch: 10,
				SessionPreFetch: 20,
				Logger:          rinq.NewLogger(true),
				PruneInterval:   20 * time.Second,
			}

			Expect(withDefaults(config)).To(Equal(config))
		})
	})
})
