package optutil_test

import (
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/internal/optutil"
)

var _ = Describe("Apply", func() {
	It("uses the correct defaults", func() {
		cfg, err := optutil.NewConfig()

		Expect(err).NotTo(HaveOccurred())
		Expect(cfg).To(Equal(optutil.Config{
			DefaultTimeout: 5 * time.Second,
			CommandWorkers: uint(runtime.GOMAXPROCS(0)),
			SessionWorkers: uint(runtime.GOMAXPROCS(0)) * 10,
			Logger:         rinq.NewLogger(false),
			PruneInterval:  3 * time.Minute,
			Product:        "",
		}))
	})
})
