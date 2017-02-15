package overpass_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/over-pass/overpass-go/src/overpass"
)

var _ = Describe("Attr", func() {
	Describe("Set", func() {
		It("returns a non-frozen attribute", func() {
			attr := overpass.Set("foo", "bar")
			expected := overpass.Attr{Key: "foo", Value: "bar"}
			Expect(attr).To(Equal(expected))
		})
	})

	Describe("Freeze", func() {
		It("returns a frozen attribute", func() {
			attr := overpass.Freeze("foo", "bar")
			expected := overpass.Attr{Key: "foo", Value: "bar", IsFrozen: true}
			Expect(attr).To(Equal(expected))
		})
	})
})
