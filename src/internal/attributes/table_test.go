package attributes_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/rinq-go/src/internal/attributes"
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("Table", func() {
	var table Table

	BeforeEach(func() {
		table = Table{
			"a": rinq.Set("a", "1"),
			"b": rinq.Set("b", "2"),
		}
	})

	Describe("Get", func() {
		It("returns the attribute", func() {
			attr, _ := table.Get("a")

			Expect(attr).To(Equal(rinq.Set("a", "1")))
		})
	})

	Describe("Each", func() {
		It("calls the function for each attribute in the table", func() {
			var attrs []rinq.Attr
			table.Each(func(attr rinq.Attr) bool {
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
			table.Each(func(attr rinq.Attr) bool {
				attrs = append(attrs, attr)
				return false
			})

			Expect(len(attrs)).To(Equal(1))
		})
	})

	Describe("IsEmpty", func() {
		It("returns true when the table is empty", func() {
			Expect(Table{}.IsEmpty()).To(BeTrue())
		})

		It("returns false when the table is not empty", func() {
			Expect(table.IsEmpty()).To(BeFalse())
		})
	})

	Describe("String", func() {
		It("returns a comma-separated string representation", func() {
			Expect(table.String()).To(SatisfyAny(
				Equal("{a=1, b=2}"),
				Equal("{b=2, a=1}"),
			))
		})
	})
})
