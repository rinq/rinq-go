package ident_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/rinq/rinq-go/src/rinq/ident"
)

var _ = Describe("MessageID", func() {
	var sessionRef = Ref{
		ID: SessionID{
			Peer: PeerID{
				Clock: 0x0123456789abcdef,
				Rand:  0x0bad,
			},
			Seq: 123,
		},
		Rev: 456,
	}

	Describe("ParseMessageID", func() {
		It("parses a human readable ID", func() {
			id, err := ParseMessageID("123456789ABCDEF-0BAD.123@456#789")

			Expect(err).ShouldNot(HaveOccurred())
			Expect(id.String()).To(Equal("123456789ABCDEF-0BAD.123@456#789"))
		})

		DescribeTable(
			"returns an error if the string is malformed",
			func(id string) {
				_, err := ParseMessageID(id)

				Expect(err).Should(HaveOccurred())
			},
			Entry("malformed", "<malformed>"),
			Entry("zero peer clock component", "0-1.1@0#1"),
			Entry("zero peer random component", "1-0.1@0#1"),
			Entry("zero message seq", "1-1.1@0#0"),
			Entry("invalid peer clock component", "x-1.1@0#1"),
			Entry("invalid peer random component", "1-x.1@0#1"),
			Entry("invalid session sequence", "1-1.x@0#1"),
			Entry("invalid session revision", "1-1.1@x#1"),
			Entry("invalid message sequence", "1-1.1@0#x"),
		)
	})

	DescribeTable(
		"Validate",
		func(subject MessageID, isValid bool) {
			if isValid {
				Expect(subject.Validate()).To(Succeed())
			} else {
				Expect(subject.Validate()).Should(HaveOccurred())
			}
		},
		Entry("zero struct", MessageID{}, false),
		Entry("zero session", MessageID{Seq: 1}, false),
		Entry("zero seq", MessageID{Ref: sessionRef}, false),
		Entry("non-zero struct", MessageID{Ref: sessionRef, Seq: 1}, true),
	)

	Describe("ShortString", func() {
		It("returns a human readable ID", func() {
			subject := MessageID{Ref: sessionRef, Seq: 789}
			Expect(subject.ShortString()).To(Equal("0BAD.123@456#789"))
		})
	})

	Describe("String", func() {
		It("returns a human readable ID", func() {
			subject := MessageID{Ref: sessionRef, Seq: 789}
			Expect(subject.String()).To(Equal("123456789ABCDEF-0BAD.123@456#789"))
		})
	})
})
