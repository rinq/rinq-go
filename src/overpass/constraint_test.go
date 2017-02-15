package overpass_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/over-pass/overpass-go/src/overpass"
)

var _ = Describe("Constraint", func() {
	Describe("String", func() {
		It("returns an asterisk when the constraint is empty", func() {
			Expect(overpass.Constraint{}.String()).To(Equal("*"))
		})

		It("it returns key value pairs in any order", func() {
			constraint := overpass.Constraint{"a": "1", "b": "2"}
			str := constraint.String()

			Expect(str).To(SatisfyAny(
				Equal("a=1, b=2"),
				Equal("b=2, a=1"),
			))
		})

		It("empty attributes are represented with an exclamation mark", func() {
			constraint := overpass.Constraint{"a": ""}
			str := constraint.String()

			Expect(str).To(Equal("!a"))
		})
	})
})
