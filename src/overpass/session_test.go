package overpass_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/over-pass/overpass-go/src/overpass"
)

var _ = Describe("NotFoundError", func() {
	Describe("Error", func() {
		It("includes the session ID", func() {
			id := overpass.SessionID{
				Peer: overpass.PeerID{Clock: 1, Rand: 2},
				Seq:  3,
			}
			err := overpass.NotFoundError{ID: id}
			Expect(err.Error()).To(Equal("session 1-0002.3 not found"))
		})
	})

	Describe("IsNotFound", func() {
		It("returns true for not found errors", func() {
			Expect(overpass.IsNotFound(overpass.NotFoundError{})).To(BeTrue())
		})

		It("returns false for other error types", func() {
			Expect(overpass.IsNotFound(errors.New(""))).To(BeFalse())
		})
	})
})
