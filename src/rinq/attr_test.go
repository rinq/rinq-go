package rinq_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/internal/attributes"
)

var _ = Describe("Attr", func() {
	Describe("String", func() {
		It("uses 'equals' syntax", func() {
			attr := rinq.Attr{Key: "foo", Value: "bar"}
			Expect(attr.String()).To(Equal("foo=bar"))
		})

		It("uses 'at' syntax for frozen attributes", func() {
			attr := rinq.Attr{Key: "foo", Value: "bar", IsFrozen: true}
			Expect(attr.String()).To(Equal("foo@bar"))
		})

		It("uses 'bang' syntax for empty frozen attributes", func() {
			attr := rinq.Attr{Key: "foo", Value: "", IsFrozen: true}
			Expect(attr.String()).To(Equal("!foo"))
		})
	})
})

var _ = Describe("Set", func() {
	It("returns a non-frozen attribute", func() {
		attr := rinq.Set("foo", "bar")
		expected := rinq.Attr{Key: "foo", Value: "bar"}
		Expect(attr).To(Equal(expected))
	})
})

var _ = Describe("Freeze", func() {
	It("returns a frozen attribute", func() {
		attr := rinq.Freeze("foo", "bar")
		expected := rinq.Attr{Key: "foo", Value: "bar", IsFrozen: true}
		Expect(attr).To(Equal(expected))
	})
})

var _ = Describe("ToMap", func() {
	It("returns a map containing the attributes", func() {
		table := attributes.Table{
			"a": rinq.Set("a", "1"),
			"b": rinq.Set("b", "2"),
		}

		Expect(rinq.ToMap(table)).To(Equal(
			map[string]rinq.Attr{
				"a": rinq.Set("a", "1"),
				"b": rinq.Set("b", "2"),
			},
		))
	})
})
