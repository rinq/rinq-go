package options_test

import (
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/options"
)

var _ = Describe("NewOptions", func() {
	It("uses the correct defaults", func() {
		opts, err := options.NewOptions()

		Expect(err).NotTo(HaveOccurred())
		Expect(opts).To(Equal(options.Options{
			DefaultTimeout: 5 * time.Second,
			CommandWorkers: uint(runtime.GOMAXPROCS(0)),
			SessionWorkers: uint(runtime.GOMAXPROCS(0)) * 10,
			Logger:         rinq.NewLogger(false),
			PruneInterval:  3 * time.Minute,
			Product:        "",
			Tracer:         opentracing.NoopTracer{},
		}))
	})
})
