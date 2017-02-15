package overpass_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/over-pass/overpass-go/src/overpass"
)

var _ = Describe("PeerID", func() {
	Describe("NewPeerID", func() {
		It("returns a valid ID", func() {
			subject := overpass.NewPeerID()
			err := subject.Validate()
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	DescribeTable(
		"Validate",
		func(subject overpass.PeerID, isValid bool) {
			if isValid {
				Expect(subject.Validate()).To(Succeed())
			} else {
				Expect(subject.Validate()).Should(HaveOccurred())
			}
		},
		Entry("zero struct", overpass.PeerID{}, false),
		Entry("zero clock component", overpass.PeerID{Rand: 1}, false),
		Entry("zero random component", overpass.PeerID{Clock: 1}, false),
		Entry("non-zero struct", overpass.PeerID{Clock: 1, Rand: 1}, true),
	)

	Describe("ShortString", func() {
		It("returns a human readable ID", func() {
			subject := overpass.PeerID{Clock: 0x0123456789abcdef, Rand: 0x0bad}
			Expect(subject.ShortString()).To(Equal("0BAD"))
		})
	})

	Describe("String", func() {
		It("returns a human readable ID", func() {
			subject := overpass.PeerID{Clock: 0x0123456789abcdef, Rand: 0x0bad}
			Expect(subject.String()).To(Equal("123456789ABCDEF-0BAD"))
		})
	})
})
