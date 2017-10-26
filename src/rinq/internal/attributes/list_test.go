package attributes_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
	. "github.com/rinq/rinq-go/src/rinq/internal/attributes"
)

var _ = Describe("List", func() {
	Describe("WriteTo", func() {
		It("writes only braces when the list is empty", func() {
			var buf bytes.Buffer

			List{}.WriteTo(&buf)

			Expect(buf.String()).To(Equal("{}"))
		})

		It("writes key/value pairs in order", func() {
			var buf bytes.Buffer
			l := List{
				rinq.Set("a", "1"),
				rinq.Set("b", "2"),
			}

			l.WriteTo(&buf)

			Expect(buf.String()).To(Equal("{a=1, b=2}"))
		})
	})

	Describe("String", func() {
		It("returns only braces when the list is empty", func() {
			Expect(List{}.String()).To(Equal("{}"))
		})

		It("returns key/value pairs in order", func() {
			l := List{
				rinq.Set("a", "1"),
				rinq.Set("b", "2"),
			}

			Expect(l.String()).To(Equal("{a=1, b=2}"))
		})
	})
})
