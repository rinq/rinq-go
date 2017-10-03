package strutil_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/rinq/rinq-go/src/rinq/internal/strutil"
)

var _ = Describe("Escape", func() {

	DescribeTable(
		"returns the expected representation",
		func(in, out string) {
			Expect(Escape(in)).To(Equal(out))
		},
		Entry("empty", "", `""`),
		Entry("no special characters", "foo", `foo`),
		Entry("escaped character", "foo\nbar", `"foo\nbar"`),
		Entry("contains space", "foo bar", `"foo bar"`),
		Entry("contains bang", "foo!bar", `"foo!bar"`),
		Entry("contains at symbol", "foo@bar", `"foo@bar"`),
		Entry("contains equal sign", "foo=bar", `"foo=bar"`),
		Entry("contains colon", "foo:bar", `"foo:bar"`),
		Entry("contains left-paren", "foo(bar", `"foo(bar"`),
		Entry("contains right-paren", "foo)bar", `"foo)bar"`),
		Entry("contains left-brace", "foo{bar", `"foo{bar"`),
		Entry("contains right-brace", "foo}bar", `"foo}bar"`),
	)
})
