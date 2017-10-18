package trace_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/rinq-go/src/rinq/trace"
)

var _ = Describe("With", func() {
	It("adds the trace ID", func() {
		ctx := With(context.Background(), "<id>")

		Expect(Get(ctx)).To(Equal("<id>"))
	})

	It("replaces an existing trace ID", func() {
		parent := With(context.Background(), "<id 1>")
		ctx := With(parent, "<id 2>")

		Expect(Get(ctx)).To(Equal("<id 2>"))
	})
})

var _ = Describe("WithRoot", func() {
	It("adds the trace ID", func() {
		ctx, added := WithRoot(context.Background(), "<id>")

		Expect(Get(ctx)).To(Equal("<id>"))
		Expect(added).To(BeTrue())
	})

	It("returns the parent context unchanged", func() {
		parent := With(context.Background(), "<id 1>")
		ctx, added := WithRoot(parent, "<id 2>")

		Expect(ctx).To(BeIdenticalTo(parent))
		Expect(added).To(BeFalse())
	})
})

var _ = Describe("Get", func() {
	It("returns an empty string when no trace ID is present", func() {
		ctx := context.Background()

		Expect(Get(ctx)).To(Equal(""))
	})
})
