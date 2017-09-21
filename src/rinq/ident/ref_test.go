package ident_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/rinq/rinq-go/src/rinq/ident"
)

var _ = Describe("Ref", func() {
	var sessionID = SessionID{
		Peer: PeerID{
			Clock: 0x0123456789abcdef,
			Rand:  0x0bad,
		},
		Seq: 123,
	}

	DescribeTable(
		"Validate",
		func(subject Ref, isValid bool) {
			if isValid {
				Expect(subject.Validate()).To(Succeed())
			} else {
				Expect(subject.Validate()).Should(HaveOccurred())
			}
		},
		Entry("zero struct", Ref{}, false),
		Entry("non-zero struct", Ref{ID: sessionID}, true),
	)

	Describe("Before", func() {
		DescribeTable(
			"compares the revision number",
			func(a, b Ref, expected bool) {
				Expect(a.Before(b)).To(Equal(expected))
			},
			Entry("a == b", SessionID{}.At(1), SessionID{}.At(1), false),
			Entry("a < b", SessionID{}.At(1), SessionID{}.At(2), true),
			Entry("a > b", SessionID{}.At(2), SessionID{}.At(1), false),
		)

		It("panics if the session IDs are not the same", func() {
			Expect(func() {
				a := SessionID{Seq: 1}.At(0)
				b := SessionID{Seq: 2}.At(0)
				a.Before(b)
			}).Should(Panic())
		})
	})

	Describe("After", func() {
		DescribeTable(
			"compares the revision number",
			func(a, b Ref, expected bool) {
				Expect(a.After(b)).To(Equal(expected))
			},
			Entry("a == b", SessionID{}.At(1), SessionID{}.At(1), false),
			Entry("a < b", SessionID{}.At(1), SessionID{}.At(2), false),
			Entry("a > b", SessionID{}.At(2), SessionID{}.At(1), true),
		)

		It("panics if the session IDs are not the same", func() {
			Expect(func() {
				a := SessionID{Seq: 1}.At(0)
				b := SessionID{Seq: 2}.At(0)
				a.After(b)
			}).Should(Panic())
		})
	})

	Describe("Message", func() {
		It("retruns a new MessageID", func() {
			subject := Ref{ID: sessionID, Rev: 456}
			messageID := subject.Message(123)
			Expect(messageID.Ref).To(Equal(subject))
			Expect(messageID.Seq).To(Equal(uint32(123)))
		})
	})

	Describe("String", func() {
		It("returns a human readable string", func() {
			subject := Ref{ID: sessionID, Rev: 456}
			Expect(subject.String()).To(Equal("123456789ABCDEF-0BAD.123@456"))
		})
	})
})
