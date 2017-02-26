package overpass_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/over-pass/overpass-go/src/overpass"
)

var _ = Describe("Command", func() {
	Describe("Failure", func() {
		Describe("Error", func() {
			It("includes both the type and the message", func() {
				err := overpass.Failure{Type: "<type>", Message: "<message>"}
				Expect(err.Error()).To(Equal("<type>: <message>"))
			})
		})
	})

	Describe("IsFailure", func() {
		It("returns true for failures", func() {
			r := overpass.IsFailure(overpass.Failure{})
			Expect(r).To(BeTrue())
		})

		It("returns false for other error types", func() {
			r := overpass.IsFailure(errors.New(""))
			Expect(r).To(BeFalse())
		})
	})

	Describe("IsFailureType", func() {
		It("returns true for failures with the same type", func() {
			r := overpass.IsFailureType("foo", overpass.Failure{Type: "foo"})
			Expect(r).To(BeTrue())
		})

		It("returns false for failures with a different type", func() {
			r := overpass.IsFailureType("foo", overpass.Failure{Type: "bar"})
			Expect(r).To(BeFalse())
		})

		It("returns false for other error types", func() {
			r := overpass.IsFailureType("foo", errors.New(""))
			Expect(r).To(BeFalse())
		})

		It("panics if the type is empty", func() {
			f := func() {
				overpass.IsFailureType("", errors.New(""))
			}
			Expect(f).Should(Panic())
		})
	})

	Describe("FailureType", func() {
		It("returns the failure type", func() {
			r := overpass.FailureType(overpass.Failure{Type: "foo"})
			Expect(r).To(Equal("foo"))
		})

		It("returns empty string for other error types", func() {
			r := overpass.FailureType(errors.New(""))
			Expect(r).To(Equal(""))
		})

		It("panics if the type is empty", func() {
			f := func() {
				overpass.FailureType(overpass.Failure{})
			}
			Expect(f).Should(Panic())
		})
	})

	Describe("IsCommandError", func() {
		It("returns true for Failure", func() {
			r := overpass.IsCommandError(overpass.Failure{Type: "foo"})
			Expect(r).To(BeTrue())
		})

		It("returns true for CommandError", func() {
			r := overpass.IsCommandError(overpass.CommandError(""))
			Expect(r).To(BeTrue())
		})

		It("returns false for other error types", func() {
			r := overpass.IsCommandError(errors.New(""))
			Expect(r).To(BeFalse())
		})
	})

	Describe("CommandError", func() {
		Describe("Error", func() {
			It("returns the message", func() {
				err := overpass.CommandError("<message>")
				Expect(err.Error()).To(Equal("<message>"))
			})

			It("returns a message for the default value", func() {
				err := overpass.CommandError("")
				Expect(err.Error()).To(Equal("unexpected command error"))
			})
		})
	})

	DescribeTable(
		"ValidateNamespace",
		func(namespace string, expected string) {
			err := overpass.ValidateNamespace(namespace)

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
