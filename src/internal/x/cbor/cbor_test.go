package cbor_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/rinq-go/src/internal/x/cbor"
)

var _ = Describe("Encode", func() {
	It("returns the expected binary representation", func() {
		var buf bytes.Buffer
		err := Encode(&buf, 123)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(buf.Bytes()).To(Equal([]byte{24, 123}))
	})
})

var _ = Describe("MustEncode", func() {
	It("returns the expected binary representation", func() {
		var buf bytes.Buffer
		MustEncode(&buf, 123)

		Expect(buf.Bytes()).To(Equal([]byte{24, 123}))
	})
})

var _ = Describe("Decode", func() {
	It("produces the expected value", func() {
		buf := bytes.NewBuffer([]byte{24, 123})

		var v interface{}
		err := Decode(buf, &v)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(v).To(Equal(uint64(123)))
	})
})

var _ = Describe("MustDecode", func() {
	It("produces the expected value", func() {
		buf := bytes.NewBuffer([]byte{24, 123})

		var v interface{}
		MustDecode(buf, &v)

		Expect(v).To(Equal(uint64(123)))
	})
})

var _ = Describe("DecodeBytes", func() {
	It("produces the expected value", func() {
		buf := []byte{24, 123}

		var v interface{}
		err := DecodeBytes(buf, &v)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(v).To(Equal(uint64(123)))
	})
})

var _ = Describe("MustDecodeBytes", func() {
	It("produces the expected value", func() {
		buf := []byte{24, 123}

		var v interface{}
		MustDecodeBytes(buf, &v)

		Expect(v).To(Equal(uint64(123)))
	})
})
