package opentr_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/internal/opentr"
)

var _ = Describe("AddTraceID", func() {
	It("adds the traceID tag to the span", func() {
		span := &mockSpan{}

		opentr.AddTraceID(span, "<id>")

		Expect(span.tags).Should(HaveKeyWithValue("traceID", "<id>"))
	})

	It("does not add the traceID tag to the span if the id is empty", func() {
		span := &mockSpan{}

		opentr.AddTraceID(span, "")

		Expect(span.tags).ShouldNot(HaveKey("traceID"))
	})
})
