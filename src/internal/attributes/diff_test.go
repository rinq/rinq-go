package attributes_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
	. "github.com/rinq/rinq-go/src/internal/attributes"
)

var _ = Describe("Diff", func() {
	Describe("NewDiff", func() {
		It("sets the namespace", func() {
			d := NewDiff("ns", 0)

			Expect(d.Namespace).To(Equal("ns"))
		})
	})

	Describe("Append", func() {
		It("appends the attributes", func() {
			d := NewDiff("ns", 0)
			attrs := VList{
				{Attr: rinq.Set("a", "1")},
				{Attr: rinq.Set("b", "2")},
			}

			d.Append(attrs...)

			Expect(d.VList).To(Equal(attrs))
		})
	})

	Describe("String", func() {
		Context("when the diff is empty", func() {
			It("returns the namespace and braces", func() {
				d := NewDiff("ns", 0)
				Expect(d.String()).To(Equal("ns::{}"))
			})
		})

		Context("when the diff contains attributes", func() {
			var d *Diff

			BeforeEach(func() {
				d = NewDiff("ns", 1)
			})

			It("renders key/value pairs in order", func() {
				d.Append(
					VAttr{Attr: rinq.Set("a", "1")},
					VAttr{Attr: rinq.Set("b", "2")},
				)

				Expect(d.String()).To(Equal("ns::{a=1, b=2}"))
			})

			It("renders new attributes with a plus-sign", func() {
				d.Append(VAttr{
					Attr:      rinq.Set("a", "1"),
					CreatedAt: 1,
				})

				Expect(d.String()).To(Equal("ns::{+a=1}"))
			})

			It("renders existing attributes normally", func() {
				d.Revision = 2
				d.Append(VAttr{
					Attr:      rinq.Set("a", "1"),
					CreatedAt: 1,
				})

				Expect(d.String()).To(Equal("ns::{a=1}"))
			})
		})
	})

	Describe("StringWithoutNamespace", func() {
		Context("when the diff is empty", func() {
			It("returns the namespace and braces", func() {
				d := NewDiff("ns", 0)
				Expect(d.StringWithoutNamespace()).To(Equal("{}"))
			})
		})

		Context("when the diff contains attributes", func() {
			var d *Diff

			BeforeEach(func() {
				d = NewDiff("ns", 1)
			})

			It("renders key/value pairs in order", func() {
				d.Append(
					VAttr{Attr: rinq.Set("a", "1")},
					VAttr{Attr: rinq.Set("b", "2")},
				)

				Expect(d.StringWithoutNamespace()).To(Equal("{a=1, b=2}"))
			})

			It("renders new attributes with a plus-sign", func() {
				d.Append(VAttr{
					Attr:      rinq.Set("a", "1"),
					CreatedAt: 1,
				})

				Expect(d.StringWithoutNamespace()).To(Equal("{+a=1}"))
			})

			It("renders existing attributes normally", func() {
				d.Revision = 2
				d.Append(VAttr{
					Attr:      rinq.Set("a", "1"),
					CreatedAt: 1,
				})

				Expect(d.StringWithoutNamespace()).To(Equal("{a=1}"))
			})
		})
	})
})
