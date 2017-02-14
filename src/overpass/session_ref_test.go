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

	Describe("String", func() {
		It("returns a human readable string", func() {
			subject := overpass.SessionRef{ID: sessionID, Rev: 456}
			Expect(subject.String()).To(Equal("123456789ABCDEF-0BAD.123@456"))
		})
	})
})
