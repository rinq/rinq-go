package rinq_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("Constraint", func() {
	Describe("String", func() {
		It("returns an asterisk when the constraint is empty", func() {
			Expect(rinq.Constraint{}.String()).To(Equal("*"))
		})

		It("it returns key value pairs in any order", func() {
			constraint := rinq.Constraint{"a": "1", "b": "2"}
			str := constraint.String()

			Expect(str).To(SatisfyAny(
				Equal("a=1, b=2"),
				Equal("b=2, a=1"),
			))
		})

		It("empty attributes are represented with an exclamation mark", func() {
			constraint := rinq.Constraint{"a": ""}
			str := constraint.String()

			Expect(str).To(Equal("!a"))
		})
	})
})
