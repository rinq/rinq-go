package attributes_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
	. "github.com/rinq/rinq-go/src/internal/attributes"
)

var _ = Describe("VList", func() {
	var list VList

	BeforeEach(func() {
		list = VList{
			{Attr: rinq.Set("a", "1")},
			{Attr: rinq.Set("b", "2")},
		}
	})

	Describe("Each", func() {
		It("calls the function for each attribute in the table", func() {
			var attrs []rinq.Attr
			list.Each(func(attr rinq.Attr) bool {
				attrs = append(attrs, attr)
				return true
			})

			Expect(attrs).To(ConsistOf(
				rinq.Set("a", "1"),
				rinq.Set("b", "2"),
			))
		})

		It("stops iteration if the function returns false", func() {
			var attrs []rinq.Attr
			list.Each(func(attr rinq.Attr) bool {
				attrs = append(attrs, attr)
				return false
			})

			Expect(len(attrs)).To(Equal(1))
		})
	})

	Describe("IsEmpty", func() {
		It("returns true when the table is empty", func() {
			Expect(VList{}.IsEmpty()).To(BeTrue())
		})

		It("returns false when the table is not empty", func() {
			Expect(list.IsEmpty()).To(BeFalse())
		})
	})

	Describe("String", func() {
		It("returns a comma-separated string representation", func() {
			Expect(list.String()).To(Equal("{a=1, b=2}"))
		})
	})
})
