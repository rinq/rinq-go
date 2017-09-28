package attrmeta_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
)

var _ = Describe("Table", func() {
	Describe("Clone", func() {
		It("returns a new table containing the same attributes", func() {
			t := attrmeta.Table{
				"a": attrmeta.Attr{Attr: rinq.Set("a", "1")},
				"b": attrmeta.Attr{Attr: rinq.Set("b", "2")},
			}

			c := t.Clone()

			Expect(c).To(Equal(t))
			Expect(c).NotTo(BeIdenticalTo(t))
		})

		It("returns a non-nil table when cloning a nil table", func() {
			var t attrmeta.Table

			c := t.Clone()

			Expect(c).To(BeEmpty())
			Expect(c).NotTo(BeNil())
		})
	})

	Describe("String", func() {
		It("returns an empty string when the table is empty", func() {
			Expect(attrmeta.Table{}.String()).To(Equal(""))
		})

		It("it returns key value pairs in any order", func() {
			t := attrmeta.Table{
				"a": attrmeta.Attr{Attr: rinq.Set("a", "1")},
				"b": attrmeta.Attr{Attr: rinq.Set("b", "2")},
			}
			str := t.String()

			Expect(str).To(SatisfyAny(
				Equal("a=1, b=2"),
				Equal("b=2, a=1"),
			))
		})
	})
})
