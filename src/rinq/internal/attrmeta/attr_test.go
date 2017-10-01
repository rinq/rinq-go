package attrmeta_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
)

var _ = Describe("Attr", func() {
	Describe("WriteTo", func() {
		buf := &bytes.Buffer{}

		BeforeEach(func() {
			buf.Reset()
		})

		It("renders attributes with an equal-sign", func() {
			attr := attrmeta.Attr{Attr: rinq.Set("foo", "bar")}

			attr.WriteTo(buf)

			Expect(buf.String()).To(Equal("foo=bar"))
		})

		It("renders frozen attributes with an at-sign", func() {
			attr := attrmeta.Attr{Attr: rinq.Freeze("foo", "bar")}

			attr.WriteTo(buf)

			Expect(buf.String()).To(Equal("foo@bar"))
		})

		It("renders empty attributes with a minus", func() {
			attr := attrmeta.Attr{Attr: rinq.Set("foo", "")}

			attr.WriteTo(buf)

			Expect(buf.String()).To(Equal("-foo"))
		})

		It("renders frozen empty attributes with a bang", func() {
			attr := attrmeta.Attr{Attr: rinq.Freeze("foo", "")}

			attr.WriteTo(buf)

			Expect(buf.String()).To(Equal("!foo"))
		})
	})
})
