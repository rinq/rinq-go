package trace_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/over-pass/overpass-go/src/internal/trace"
)

var _ = Describe("Context", func() {
	Describe("WithID and GetID", func() {
		It("transports the trace ID in the context", func() {
			ctx := trace.With(context.Background(), "<id>")

			Expect(trace.Get(ctx)).To(Equal("<id>"))
		})

		It("returns an empty string when no trace ID is present", func() {
			ctx := context.Background()

			Expect(trace.Get(ctx)).To(Equal(""))
		})
	})
})
