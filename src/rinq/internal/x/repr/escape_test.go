package repr_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq/internal/x/repr"
)

var _ = Describe("Escape", func() {
	DescribeTable(
		"returns the expected representation",
		func(in, out string) {
			Expect(repr.Escape(in)).To(Equal(out))
		},
		Entry("empty", "", `""`),
		Entry("no special characters", "Aa0_-", `Aa0_-`),
		Entry("special characters", "foo bar", `"foo bar"`),
	)
})
