package overpass_test

import (
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/over-pass/overpass-go/src/overpass"
)

var _ = Describe("DefaultConfig", func() {
	It("uses the correct defaults", func() {
		Expect(overpass.DefaultConfig).To(Equal(overpass.Config{
			DefaultTimeout:  5 * time.Second,
			CommandPreFetch: runtime.GOMAXPROCS(0),
			SessionPreFetch: runtime.GOMAXPROCS(0) * 10,
			Logger:          overpass.NewLogger(false),
			PruneInterval:   3 * time.Minute,
		}))
	})
})
