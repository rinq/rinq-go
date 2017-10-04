package nsutil_test

import (
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq/internal/nsutil"
)

var _ = DescribeTable(
	"ValidateNamespace",
	func(namespace string, expected string) {
		err := nsutil.Validate(namespace)

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
