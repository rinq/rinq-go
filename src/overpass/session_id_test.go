package overpass_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/over-pass/overpass-go/src/overpass"
)

var _ = Describe("SessionID", func() {
	peerID := overpass.PeerID{
		Clock: 0x0123456789abcdef,
		Rand:  0x0bad,
	}

	Describe("ParseSessionID", func() {
		It("parses a human readable ID", func() {
			id, err := overpass.ParseSessionID("123456789ABCDEF-0BAD.123")

			Expect(err).ShouldNot(HaveOccurred())
			Expect(id.String()).To(Equal("123456789ABCDEF-0BAD.123"))
		})

		DescribeTable(
			"returns an error if the string is malformed",
			func(id string) {
				_, err := overpass.ParseSessionID(id)

				Expect(err).Should(HaveOccurred())
			},
			Entry("malformed", "<malformed>"),
			Entry("zero peer clock component", "0-1"),
			Entry("zero peer random component", "1-0.1"),
			Entry("invalid peer clock component", "x-1.1"),
			Entry("invalid peer random component", "1-x.1"),
			Entry("invalid session sequence", "1-1.x"),
		)
	})

	DescribeTable(
		"Validate",
		func(subject overpass.SessionID, isValid bool) {
			if isValid {
				Expect(subject.Validate()).To(Succeed())
			} else {
				Expect(subject.Validate()).Should(HaveOccurred())
			}
		},
		Entry("zero struct", overpass.SessionID{}, false),
		Entry("zero peer", overpass.SessionID{Seq: 1}, false),
		Entry("non-zero struct", overpass.SessionID{Peer: peerID, Seq: 1}, true),
	)

	Describe("At", func() {
		It("creates a ref at the given version", func() {
			subject := overpass.SessionID{Peer: peerID, Seq: 123}
			ref := subject.At(456)
			Expect(ref).To(Equal(overpass.SessionRef{ID: subject, Rev: 456}))
		})
	})

	Describe("ShortString", func() {
		It("returns a human readable ID", func() {
			subject := overpass.SessionID{Peer: peerID, Seq: 123}
			Expect(subject.ShortString()).To(Equal("0BAD.123"))
		})
	})

	Describe("String", func() {
		It("returns a human readable ID", func() {
			subject := overpass.SessionID{Peer: peerID, Seq: 123}
			Expect(subject.String()).To(Equal("123456789ABCDEF-0BAD.123"))
		})
	})
})
