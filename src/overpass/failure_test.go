package overpass_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/over-pass/overpass-go/src/overpass"
)

var _ = Describe("Failure", func() {
	Describe("Error", func() {
		It("includes both the type and the message", func() {
			f := overpass.Failure{Type: "<type>", Message: "<message>"}
			Expect(f.Error()).To(Equal("<type>: <message>"))
		})
	})

	Describe("IsFailure", func() {
		It("returns true for failures", func() {
			Expect(overpass.IsFailure(overpass.Failure{})).To(BeTrue())
		})

		It("returns false for other error types", func() {
			Expect(overpass.IsFailure(errors.New(""))).To(BeFalse())
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
		It("returns the type of failures", func() {
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
})
