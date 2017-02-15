package overpass_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/over-pass/overpass-go/src/overpass"
)

var _ = Describe("SessionRef", func() {
	var sessionID = overpass.SessionID{
		Peer: overpass.PeerID{
			Clock: 0x0123456789abcdef,
			Rand:  0x0bad,
		},
		Seq: 123,
	}

	DescribeTable(
		"Validate",
		func(subject overpass.SessionRef, isValid bool) {
			if isValid {
				Expect(subject.Validate()).To(Succeed())
			} else {
				Expect(subject.Validate()).Should(HaveOccurred())
			}
		},
		Entry("zero struct", overpass.SessionRef{}, false),
		Entry("non-zero struct", overpass.SessionRef{ID: sessionID}, true),
	)

	Describe("Before", func() {
		DescribeTable(
			"compares the revision number",
			func(a, b overpass.SessionRef, expected bool) {
				Expect(a.Before(b)).To(Equal(expected))
			},
			Entry("a == b", overpass.SessionID{}.At(1), overpass.SessionID{}.At(1), false),
			Entry("a < b", overpass.SessionID{}.At(1), overpass.SessionID{}.At(2), true),
			Entry("a > b", overpass.SessionID{}.At(2), overpass.SessionID{}.At(1), false),
		)

		It("panics if the session IDs are not the same", func() {
			Expect(func() {
				a := overpass.SessionID{Seq: 1}.At(0)
				b := overpass.SessionID{Seq: 2}.At(0)
				a.Before(b)
			}).Should(Panic())
		})
	})

	Describe("After", func() {
		DescribeTable(
			"compares the revision number",
			func(a, b overpass.SessionRef, expected bool) {
				Expect(a.After(b)).To(Equal(expected))
			},
			Entry("a == b", overpass.SessionID{}.At(1), overpass.SessionID{}.At(1), false),
			Entry("a < b", overpass.SessionID{}.At(1), overpass.SessionID{}.At(2), false),
			Entry("a > b", overpass.SessionID{}.At(2), overpass.SessionID{}.At(1), true),
		)

		It("panics if the session IDs are not the same", func() {
			Expect(func() {
				a := overpass.SessionID{Seq: 1}.At(0)
				b := overpass.SessionID{Seq: 2}.At(0)
				a.After(b)
			}).Should(Panic())
		})
	})

	Describe("String", func() {
		It("returns a human readable string", func() {
			subject := overpass.SessionRef{ID: sessionID, Rev: 456}
			Expect(subject.String()).To(Equal("123456789ABCDEF-0BAD.123@456"))
		})
	})
})
