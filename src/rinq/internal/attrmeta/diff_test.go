package attrmeta_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
)

var _ = Describe("Diff", func() {
	Describe("NewDiff", func() {
		It("sets the namespace", func() {
			d := attrmeta.NewDiff("ns", 0, 0)

			Expect(d.Namespace).To(Equal("ns"))
		})

		It("reserves capacity for the given number of attributes", func() {
			d := attrmeta.NewDiff("ns", 0, 3)

			Expect(cap(d.Attrs)).To(Equal(3))
		})
	})

	Describe("Append", func() {
		It("appends the attributes", func() {
			d := attrmeta.NewDiff("ns", 0, 0)
			attrs := attrmeta.List{
				{Attr: rinq.Set("a", "1")},
				{Attr: rinq.Set("b", "2")},
			}

			d.Append(attrs...)

			Expect(d.Attrs).To(Equal(attrs))
		})
	})

	Describe("IsEmpty", func() {
		It("returns true if the diff is empty", func() {
			d := attrmeta.NewDiff("ns", 0, 0)

			Expect(d.IsEmpty()).To(BeTrue())
		})

		It("returns false if the diff contains attributes", func() {
			d := attrmeta.NewDiff("ns", 0, 0)
			d.Append(attrmeta.Attr{})

			Expect(d.IsEmpty()).To(BeFalse())
		})
	})

	Describe("WriteTo", func() {
		buf := &bytes.Buffer{}

		BeforeEach(func() {
			buf.Reset()
		})

		Context("when the diff is empty", func() {
			d := attrmeta.NewDiff("ns", 0, 0)

			It("writes the namespace and braces", func() {
				d.WriteTo(buf)

				Expect(buf.String()).To(Equal("ns::{}"))
			})
		})

		Context("when the diff contains attributes", func() {
			var d *attrmeta.Diff

			BeforeEach(func() {
				d = attrmeta.NewDiff("ns", 1, 0)
			})

			It("writes key/value pairs in order", func() {
				d.Append(
					attrmeta.Attr{Attr: rinq.Set("a", "1")},
					attrmeta.Attr{Attr: rinq.Set("b", "2")},
				)

				d.WriteTo(buf)

				Expect(buf.String()).To(SatisfyAny(
					Equal("ns::{a=1, b=2}"),
					Equal("ns::{b=2, a=1}"),
				))
			})

			It("renders new attributes with a plus-sign", func() {
				d.Append(attrmeta.Attr{
					Attr:      rinq.Set("a", "1"),
					CreatedAt: 1,
				})

				d.WriteTo(buf)

				Expect(buf.String()).To(Equal("ns::{+a=1}"))
			})

			It("renders existing attributes normally", func() {
				d.Revision = 2
				d.Append(attrmeta.Attr{
					Attr:      rinq.Set("a", "1"),
					CreatedAt: 1,
				})

				d.WriteTo(buf)

				Expect(buf.String()).To(Equal("ns::{a=1}"))
			})
		})
	})

	Describe("WriteWithoutNamespaceTo", func() {
		buf := &bytes.Buffer{}

		BeforeEach(func() {
			buf.Reset()
		})

		Context("when the diff is empty", func() {
			d := attrmeta.NewDiff("ns", 0, 0)

			It("writes the namespace and braces", func() {
				d.WriteWithoutNamespaceTo(buf)

				Expect(buf.String()).To(Equal("{}"))
			})
		})

		Context("when the diff contains attributes", func() {
			var d *attrmeta.Diff

			BeforeEach(func() {
				d = attrmeta.NewDiff("ns", 1, 0)
			})

			It("writes key/value pairs in order", func() {
				d.Append(
					attrmeta.Attr{Attr: rinq.Set("a", "1")},
					attrmeta.Attr{Attr: rinq.Set("b", "2")},
				)

				d.WriteWithoutNamespaceTo(buf)

				Expect(buf.String()).To(SatisfyAny(
					Equal("{a=1, b=2}"),
					Equal("{b=2, a=1}"),
				))
			})

			It("renders new attributes with a plus-sign", func() {
				d.Append(attrmeta.Attr{
					Attr:      rinq.Set("a", "1"),
					CreatedAt: 1,
				})

				d.WriteWithoutNamespaceTo(buf)

				Expect(buf.String()).To(Equal("{+a=1}"))
			})

			It("renders existing attributes normally", func() {
				d.Revision = 2
				d.Append(attrmeta.Attr{
					Attr:      rinq.Set("a", "1"),
					CreatedAt: 1,
				})

				d.WriteWithoutNamespaceTo(buf)

				Expect(buf.String()).To(Equal("{a=1}"))
			})
		})
	})

	Describe("String", func() {
		It("returns only the namespace and braces when the diff is empty", func() {
			Expect(attrmeta.NewDiff("ns", 0, 0).String()).To(Equal("ns::{}"))
		})

		It("returns key/value pairs in any order", func() {
			d := attrmeta.NewDiff("ns", 1, 0)
			d.Append(
				attrmeta.Attr{Attr: rinq.Set("a", "1")},
				attrmeta.Attr{Attr: rinq.Set("b", "2")},
			)

			Expect(d.String()).To(SatisfyAny(
				Equal("ns::{a=1, b=2}"),
				Equal("ns::{b=2, a=1}"),
			))
		})
	})

	Describe("String", func() {
		It("returns only tbraces when the diff is empty", func() {
			Expect(attrmeta.NewDiff("ns", 0, 0).StringWithoutNamespace()).To(Equal("{}"))
		})

		It("returns key/value pairs in any order", func() {
			d := attrmeta.NewDiff("ns", 1, 0)
			d.Append(
				attrmeta.Attr{Attr: rinq.Set("a", "1")},
				attrmeta.Attr{Attr: rinq.Set("b", "2")},
			)

			Expect(d.StringWithoutNamespace()).To(SatisfyAny(
				Equal("{a=1, b=2}"),
				Equal("{b=2, a=1}"),
			))
		})
	})
})
