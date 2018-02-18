package amqpx_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/rinq-go/src/rinqamqp/internal/amqpx"
	"github.com/streadway/amqp"
)

var _ = Describe("SetHeader", func() {
	It("sets the header if the header map is nil", func() {
		msg := &amqp.Publishing{}

		SetHeader(msg, "a", 1)

		Expect(msg.Headers["a"]).To(Equal(1))
	})

	It("adds a new header", func() {
		msg := &amqp.Publishing{
			Headers: amqp.Table{
				"a": 1,
			},
		}

		SetHeader(msg, "b", 2)

		Expect(msg.Headers).To(Equal(amqp.Table{
			"a": 1,
			"b": 2,
		}))
	})
})

var _ = Describe("GetHeaderString", func() {
	It("returns the header if it is a string", func() {
		msg := &amqp.Delivery{
			Headers: amqp.Table{
				"a": "1",
			},
		}

		v, err := GetHeaderString(msg, "a")

		Expect(err).ShouldNot(HaveOccurred())
		Expect(v).To(Equal("1"))
	})

	It("returns an error if the header is not set", func() {
		msg := &amqp.Delivery{}

		_, err := GetHeaderString(msg, "a")

		Expect(err).Should(HaveOccurred())
	})

	It("returns an error if the header is not a string", func() {
		msg := &amqp.Delivery{
			Headers: amqp.Table{
				"a": 1,
			},
		}

		_, err := GetHeaderString(msg, "a")

		Expect(err).Should(HaveOccurred())
	})
})

var _ = Describe("GetHeaderBytes", func() {
	It("returns the header if it is a string", func() {
		msg := &amqp.Delivery{
			Headers: amqp.Table{
				"a": []byte{'1'},
			},
		}

		v, err := GetHeaderBytes(msg, "a")

		Expect(err).ShouldNot(HaveOccurred())
		Expect(v).To(Equal([]byte{'1'}))
	})

	It("returns an error if the header is not set", func() {
		msg := &amqp.Delivery{}

		_, err := GetHeaderBytes(msg, "a")

		Expect(err).Should(HaveOccurred())
	})

	It("returns an error if the header is not a string", func() {
		msg := &amqp.Delivery{
			Headers: amqp.Table{
				"a": 1,
			},
		}

		_, err := GetHeaderBytes(msg, "a")

		Expect(err).Should(HaveOccurred())
	})
})

var _ = Describe("GetHeaderBytesOptional", func() {
	It("returns the header if it is a string", func() {
		msg := &amqp.Delivery{
			Headers: amqp.Table{
				"a": []byte{'1'},
			},
		}

		v, ok, err := GetHeaderBytesOptional(msg, "a")

		Expect(err).ShouldNot(HaveOccurred())
		Expect(v).To(Equal([]byte{'1'}))
		Expect(ok).To(BeTrue())
	})

	It("returns false if the header is not set", func() {
		msg := &amqp.Delivery{}

		_, ok, err := GetHeaderBytesOptional(msg, "a")

		Expect(err).ShouldNot(HaveOccurred())
		Expect(ok).To(BeFalse())
	})

	It("returns an error if the header is not a string", func() {
		msg := &amqp.Delivery{
			Headers: amqp.Table{
				"a": 1,
			},
		}

		_, _, err := GetHeaderBytesOptional(msg, "a")

		Expect(err).Should(HaveOccurred())
	})
})
