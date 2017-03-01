package rinq_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("NotFoundError", func() {
	Describe("Error", func() {
		It("includes the session ID", func() {
			id := rinq.SessionID{
				Peer: rinq.PeerID{Clock: 1, Rand: 2},
				Seq:  3,
			}
			err := rinq.NotFoundError{ID: id}
			Expect(err.Error()).To(Equal("session 1-0002.3 not found"))
		})
	})

	Describe("IsNotFound", func() {
		It("returns true for not found errors", func() {
			Expect(rinq.IsNotFound(rinq.NotFoundError{})).To(BeTrue())
		})

		It("returns false for other error types", func() {
			Expect(rinq.IsNotFound(errors.New(""))).To(BeFalse())
		})
	})
})
