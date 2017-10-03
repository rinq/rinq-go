package attrmeta_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
)

var _ = Describe("Namespace", func() {
	Describe("Clone", func() {
		var ns attrmeta.Namespace

		BeforeEach(func() {
			ns = attrmeta.Namespace{
				"a": {Attr: rinq.Set("a", "1")},
				"b": {Attr: rinq.Set("b", "2")},
			}
		})

		It("returns a different instance", func() {
			c := ns.Clone()
			c["c"] = attrmeta.Attr{Attr: rinq.Set("c", "3")}

			Expect(ns).NotTo(HaveKey("c"))
		})

		It("contains the same attributes", func() {
			c := ns.Clone()

			Expect(c).To(Equal(ns))
		})

		It("returns a non-nil namespace when cloning a nil namespace", func() {
			ns = nil
			c := ns.Clone()

			Expect(c).To(BeEmpty())
			Expect(c).NotTo(BeNil())
		})
	})

	Describe("WriteTo", func() {
		buf := &bytes.Buffer{}

		BeforeEach(func() {
			buf.Reset()
		})

		Context("when the namespace is empty", func() {
			ns := attrmeta.Namespace{}

			It("returns false", func() {
				Expect(ns.WriteTo(buf)).To(BeFalse())
			})

			It("writes only braces", func() {
				ns.WriteTo(buf)

				Expect(buf.String()).To(Equal("{}"))
			})
		})

		Context("when the namespace is not empty", func() {
			ns := attrmeta.Namespace{
				"a": {Attr: rinq.Set("a", "1")},
				"b": {Attr: rinq.Set("b", "2")},
			}

			It("returns true", func() {
				Expect(ns.WriteTo(buf)).To(BeTrue())
			})

			It("writes key/value pairs in any order", func() {
				var buf bytes.Buffer

				ns.WriteTo(&buf)

				Expect(buf.String()).To(SatisfyAny(
					Equal("{a=1, b=2}"),
					Equal("{b=2, a=1}"),
				))
			})
		})

		It("excludes non-frozen attributes with empty values", func() {
			var buf bytes.Buffer
			ns := attrmeta.Namespace{
				"a": {Attr: rinq.Freeze("a", "")},
				"b": {Attr: rinq.Set("b", "")},
			}

			ns.WriteTo(&buf)

			Expect(buf.String()).To(Equal("{!a}"))
		})
	})

	Describe("WriteWithNameTo", func() {
		buf := &bytes.Buffer{}

		BeforeEach(func() {
			buf.Reset()
		})

		Context("when the namespace is empty", func() {
			ns := attrmeta.Namespace{}

			It("returns false", func() {
				Expect(ns.WriteWithNameTo(buf, "<name>")).To(BeFalse())
			})

			It("writes the name and braces", func() {
				ns.WriteWithNameTo(buf, "<name>")

				Expect(buf.String()).To(Equal("<name>::{}"))
			})
		})

		Context("when the namespace is not empty", func() {
			ns := attrmeta.Namespace{
				"a": {Attr: rinq.Set("a", "1")},
				"b": {Attr: rinq.Set("b", "2")},
			}

			It("returns true", func() {
				Expect(ns.WriteWithNameTo(buf, "<name>")).To(BeTrue())
			})

			It("writes key/value pairs in any order", func() {
				ns := attrmeta.Namespace{
					"a": {Attr: rinq.Set("a", "1")},
					"b": {Attr: rinq.Set("b", "2")},
				}

				ns.WriteWithNameTo(buf, "<name>")

				Expect(buf.String()).To(SatisfyAny(
					Equal("<name>::{a=1, b=2}"),
					Equal("<name>::{b=2, a=1}"),
				))
			})
		})
	})

	Describe("String", func() {
		It("returns only braces when the namespace is empty", func() {
			Expect(attrmeta.Namespace{}.String()).To(Equal("{}"))
		})

		It("returns key/value pairs in any order", func() {
			t := attrmeta.Namespace{
				"a": {Attr: rinq.Set("a", "1")},
				"b": {Attr: rinq.Set("b", "2")},
			}
			str := t.String()

			Expect(str).To(SatisfyAny(
				Equal("{a=1, b=2}"),
				Equal("{b=2, a=1}"),
			))
		})
	})
})
