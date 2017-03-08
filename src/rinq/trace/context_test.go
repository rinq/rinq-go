package trace_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/rinq-go/src/rinq/trace"
)

var _ = Describe("With and Get", func() {
	It("transports the trace ID in the context", func() {
		ctx := With(context.Background(), "<id>")

		Expect(Get(ctx)).To(Equal("<id>"))
	})

	It("returns an empty string when no trace ID is present", func() {
		ctx := context.Background()

		Expect(Get(ctx)).To(Equal(""))
	})
})
