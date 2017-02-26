package bufferpool_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/over-pass/overpass-go/src/overpass/internal/bufferpool"
)

var _ = Describe("bufferpool", func() {
	Describe("Get", func() {
		It("returns a bytes.Buffer pointer", func() {
			buffer := bufferpool.Get()
			Expect(buffer).ShouldNot(BeNil())
		})

		It("recycles buffers", func() {
			buffer := bufferpool.Get()
			bufferpool.Put(buffer)

			Expect(bufferpool.Get()).To(Equal(buffer))
		})
	})

	Describe("Put", func() {
		It("accepts a buffer pointer", func() {
			var buffer bytes.Buffer
			bufferpool.Put(&buffer)
		})

		It("accepts a nil pointer", func() {
			var buffer *bytes.Buffer
			bufferpool.Put(buffer)

			Expect(bufferpool.Get()).ShouldNot(BeNil())
		})
	})
})
