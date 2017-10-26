package attributes_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
	. "github.com/rinq/rinq-go/src/internal/attributes"
)

var _ = Describe("VTable", func() {
	var table VTable

	BeforeEach(func() {
		table = VTable{
			"a": {Attr: rinq.Set("a", "1")},
			"b": {Attr: rinq.Set("b", "2")},
		}
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

	Describe("Clone", func() {
		It("returns a different instance", func() {
			t := table.Clone()
			t["c"] = VAttr{Attr: rinq.Set("c", "3")}

			Expect(table).NotTo(HaveKey("c"))
		})

		It("contains the same attributes", func() {
			t := table.Clone()

			Expect(t).To(Equal(table))
		})

		It("returns a non-nil table when cloning a nil table", func() {
			table = nil
			t := table.Clone()

			Expect(t).To(BeEmpty())
			Expect(t).NotTo(BeNil())
		})
	})
})
