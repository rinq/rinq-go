package attributes_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/rinq-go/src/internal/attributes"
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("ToMap", func() {
	It("returns a map containing the attributes", func() {
		table := Table{
			"a": rinq.Set("a", "1"),
			"b": rinq.Set("b", "2"),
		}

		Expect(ToMap(table)).To(Equal(
			map[string]rinq.Attr{
				"a": rinq.Set("a", "1"),
				"b": rinq.Set("b", "2"),
			},
		))
	})
})

// Describe("WriteTo", func() {
// 	It("writes only braces when the list is empty", func() {
// 		var buf bytes.Buffer
//
// 		List{}.WriteTo(&buf)
//
// 		Expect(buf.String()).To(Equal("{}"))
// 	})
//
// 	It("writes key/value pairs in order", func() {
// 		var buf bytes.Buffer
// 		l := List{
// 			rinq.Set("a", "1"),
// 			rinq.Set("b", "2"),
// 		}
//
// 		l.WriteTo(&buf)
//
// 		Expect(buf.String()).To(Equal("{a=1, b=2}"))
// 	})
// })
//
// Describe("String", func() {
// 	It("returns only braces when the list is empty", func() {
// 		Expect(List{}.String()).To(Equal("{}"))
// 	})
//
// 	It("returns key/value pairs in order", func() {
// 		l := List{
// 			rinq.Set("a", "1"),
// 			rinq.Set("b", "2"),
// 		}
//
// 		Expect(l.String()).To(Equal("{a=1, b=2}"))
// 	})
// })
