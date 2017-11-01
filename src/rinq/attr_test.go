package rinq_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
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

		It("uses 'minus' syntax for empty attributes", func() {
			attr := rinq.Attr{Key: "foo", Value: ""}
			Expect(attr.String()).To(Equal("-foo"))
		})

		It("uses 'bang' syntax for empty frozen attributes", func() {
			attr := rinq.Attr{Key: "foo", Value: "", IsFrozen: true}
			Expect(attr.String()).To(Equal("!foo"))
		})

		It("escapes attributes that contain certain characters", func() {
			attr := rinq.Attr{Key: "foo key", Value: "bar value"}
			Expect(attr.String()).To(Equal("\"foo key\"=\"bar value\""))
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
