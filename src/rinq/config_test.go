package rinq_test

import (
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("DefaultConfig", func() {
	It("uses the correct defaults", func() {
		Expect(rinq.DefaultConfig).To(Equal(rinq.Config{
			DefaultTimeout: 5 * time.Second,
			CommandWorkers: runtime.GOMAXPROCS(0),
			SessionWorkers: runtime.GOMAXPROCS(0) * 10,
			Logger:         rinq.NewLogger(false),
			PruneInterval:  3 * time.Minute,
			Product:        "",
		}))
	})
})
