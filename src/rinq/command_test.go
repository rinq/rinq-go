package rinq_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("Failure", func() {
	Describe("Error", func() {
		It("includes both the type and the message", func() {
			err := rinq.Failure{Type: "<type>", Message: "<message>"}
			Expect(err.Error()).To(Equal("<type>: <message>"))
		})
	})
})

var _ = Describe("IsFailure", func() {
	It("returns true for failures", func() {
		r := rinq.IsFailure(rinq.Failure{})
		Expect(r).To(BeTrue())
	})

	It("returns false for other error types", func() {
		r := rinq.IsFailure(errors.New(""))
		Expect(r).To(BeFalse())
	})
})

var _ = Describe("IsFailureType", func() {
	It("returns true for failures with the same type", func() {
		r := rinq.IsFailureType("foo", rinq.Failure{Type: "foo"})
		Expect(r).To(BeTrue())
	})

	It("returns false for failures with a different type", func() {
		r := rinq.IsFailureType("foo", rinq.Failure{Type: "bar"})
		Expect(r).To(BeFalse())
	})

	It("returns false for other error types", func() {
		r := rinq.IsFailureType("foo", errors.New(""))
		Expect(r).To(BeFalse())
	})

	It("panics if the type is empty", func() {
		f := func() {
			rinq.IsFailureType("", errors.New(""))
		}
		Expect(f).Should(Panic())
	})
})

var _ = Describe("FailureType", func() {
	It("returns the failure type", func() {
		r := rinq.FailureType(rinq.Failure{Type: "foo"})
		Expect(r).To(Equal("foo"))
	})

	It("returns empty string for other error types", func() {
		r := rinq.FailureType(errors.New(""))
		Expect(r).To(Equal(""))
	})

	It("panics if the type is empty", func() {
		f := func() {
			rinq.FailureType(rinq.Failure{})
		}
		Expect(f).Should(Panic())
	})
})

var _ = Describe("IsCommandError", func() {
	It("returns true for Failure", func() {
		r := rinq.IsCommandError(rinq.Failure{Type: "foo"})
		Expect(r).To(BeTrue())
	})

	It("returns true for CommandError", func() {
		r := rinq.IsCommandError(rinq.CommandError(""))
		Expect(r).To(BeTrue())
	})

	It("returns false for other error types", func() {
		r := rinq.IsCommandError(errors.New(""))
		Expect(r).To(BeFalse())
	})
})

var _ = Describe("CommandError", func() {
	Describe("Error", func() {
		It("returns the message", func() {
			err := rinq.CommandError("<message>")
			Expect(err.Error()).To(Equal("<message>"))
		})

		It("returns a message for the default value", func() {
			err := rinq.CommandError("")
			Expect(err.Error()).To(Equal("unexpected command error"))
		})
	})
})

var _ = DescribeTable(
	"ValidateNamespace",
	func(namespace string, expected string) {
		err := rinq.ValidateNamespace(namespace)

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
