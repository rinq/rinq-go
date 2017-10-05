package rinq_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

var _ = Describe("Revision", func() {
	var sessionRef = ident.Ref{
		ID: ident.SessionID{
			Peer: ident.PeerID{
				Clock: 1,
				Rand:  2,
			},
			Seq: 3,
		},
		Rev: 4,
	}

	Describe("ShouldRetry", func() {
		It("returns true for StaleFetchError", func() {
			r := rinq.ShouldRetry(rinq.StaleFetchError{})
			Expect(r).To(BeTrue())
		})

		It("returns true for StaleUpdateError", func() {
			r := rinq.ShouldRetry(rinq.StaleUpdateError{})
			Expect(r).To(BeTrue())
		})

		It("returns false for other error types", func() {
			r := rinq.ShouldRetry(rinq.FrozenAttributesError{})
			Expect(r).To(BeFalse())
		})
	})

	Describe("StaleFetchError", func() {
		Describe("Error", func() {
			It("returns the message", func() {
				err := rinq.StaleFetchError{Ref: sessionRef}
				Expect(err.Error()).To(Equal(
					"can not fetch attributes at 1-0002.3@4, one or more attributes have been modified since that revision",
				))
			})
		})
	})

	Describe("StaleUpdateError", func() {
		Describe("Error", func() {
			It("returns the message", func() {
				err := rinq.StaleUpdateError{Ref: sessionRef}
				Expect(err.Error()).To(Equal(
					"can not update or close 1-0002.3@4, the session has been modified since that revision",
				))
			})
		})
	})

	Describe("FrozenAttributesError", func() {
		Describe("Error", func() {
			It("returns the message", func() {
				err := rinq.FrozenAttributesError{Ref: sessionRef}
				Expect(err.Error()).To(Equal(
					"can not update 1-0002.3@4, the change affects one or more frozen attributes",
				))
			})
		})
	})
})
