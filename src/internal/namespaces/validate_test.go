package namespaces_test

import (
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/internal/namespaces"
)

var entries = []TableEntry{
	Entry("all valid characters", ":Aa3-_.", ""),
	Entry("typical style", "foo.bar.v1", ""),
	Entry("empty", "", "namespace must not be empty"),
	Entry("underscore", "_", "namespace '_' is reserved"),
	Entry("leading underscore", "_foo", "namespace '_foo' is reserved"),
	Entry("invalid characters", "foo bar", "namespace 'foo bar' contains invalid characters"),
}

var _ = DescribeTable(
	"Validate",
	func(namespace string, expected string) {
		err := namespaces.Validate(namespace)

		if expected == "" {
			Expect(err).ShouldNot(HaveOccurred())
		} else {
			Expect(err.Error()).To(Equal(expected))
		}
	},
	entries...,
)

var _ = DescribeTable(
	"MustValidate",
	func(namespace string, expected string) {
		fn := func() {
			namespaces.MustValidate(namespace)
		}

		if expected == "" {
			Expect(fn).ShouldNot(Panic())
		} else {
			Expect(fn).Should(Panic())
		}
	},
	entries...,
)
