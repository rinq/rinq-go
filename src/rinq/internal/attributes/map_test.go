package attributes_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
	. "github.com/rinq/rinq-go/src/rinq/internal/attributes"
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
