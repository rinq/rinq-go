package ident_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/rinq-go/src/rinq/ident"
)

var _ = Describe("MustValidate", func() {
	It("panics if the value is invalid", func() {
		Expect(func() {
			MustValidate(PeerID{})
		}).To(Panic())
	})

	It("does not panic if the value is valid", func() {
		Expect(func() {
			MustValidate(NewPeerID())
		}).NotTo(Panic())
	})
})
