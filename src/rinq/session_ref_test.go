package rinq_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("SessionRef", func() {
	var sessionID = rinq.SessionID{
		Peer: rinq.PeerID{
			Clock: 0x0123456789abcdef,
			Rand:  0x0bad,
		},
		Seq: 123,
	}

	DescribeTable(
		"Validate",
		func(subject rinq.SessionRef, isValid bool) {
			if isValid {
				Expect(subject.Validate()).To(Succeed())
			} else {
				Expect(subject.Validate()).Should(HaveOccurred())
			}
		},
		Entry("zero struct", rinq.SessionRef{}, false),
		Entry("non-zero struct", rinq.SessionRef{ID: sessionID}, true),
	)

	Describe("Before", func() {
		DescribeTable(
			"compares the revision number",
			func(a, b rinq.SessionRef, expected bool) {
				Expect(a.Before(b)).To(Equal(expected))
			},
			Entry("a == b", rinq.SessionID{}.At(1), rinq.SessionID{}.At(1), false),
			Entry("a < b", rinq.SessionID{}.At(1), rinq.SessionID{}.At(2), true),
			Entry("a > b", rinq.SessionID{}.At(2), rinq.SessionID{}.At(1), false),
		)

		It("panics if the session IDs are not the same", func() {
			Expect(func() {
				a := rinq.SessionID{Seq: 1}.At(0)
				b := rinq.SessionID{Seq: 2}.At(0)
				a.Before(b)
			}).Should(Panic())
		})
	})

	Describe("After", func() {
		DescribeTable(
			"compares the revision number",
			func(a, b rinq.SessionRef, expected bool) {
				Expect(a.After(b)).To(Equal(expected))
			},
			Entry("a == b", rinq.SessionID{}.At(1), rinq.SessionID{}.At(1), false),
			Entry("a < b", rinq.SessionID{}.At(1), rinq.SessionID{}.At(2), false),
			Entry("a > b", rinq.SessionID{}.At(2), rinq.SessionID{}.At(1), true),
		)

		It("panics if the session IDs are not the same", func() {
			Expect(func() {
				a := rinq.SessionID{Seq: 1}.At(0)
				b := rinq.SessionID{Seq: 2}.At(0)
				a.After(b)
			}).Should(Panic())
		})
	})

	Describe("String", func() {
		It("returns a human readable string", func() {
			subject := rinq.SessionRef{ID: sessionID, Rev: 456}
			Expect(subject.String()).To(Equal("123456789ABCDEF-0BAD.123@456"))
		})
	})
})
