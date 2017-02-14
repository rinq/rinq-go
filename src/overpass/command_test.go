package overpass_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/over-pass/overpass-go/src/overpass"
)

var _ = Describe("Command", func() {
	DescribeTable(
		"ValidateNamespace",
		func(namespace string, expected string) {
			err := overpass.IsValidNamespace(namespace)

			if expected == "" {
				Expect(err).ShouldNot(HaveOccurred())
			} else {
				Expect(err.Error()).To(Equal(expected))
			}
		},
		Entry("all valid characters", ":Aa3-_.", ""),
		Entry("typical style", "foo.bar.v1", ""),
		Entry("empty", "", "namespace must not be empty"),
		Entry("underscore", "_", "namespace '_' is reserved"),
		Entry("leading underscore", "_foo", "namespace '_foo' is reserved"),
		Entry("invalid characters", "foo bar", "namespace 'foo bar' contains invalid characters"),
	)
})
