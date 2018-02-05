package marshaling_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rinq/rinq-go/src/rinqamqp/internal/refactor/marshaling"
	"github.com/streadway/amqp"
)

var _ = Describe("Trace", func() {
	Describe("PackTrace", func() {
		It("sets the correlation ID", func() {
			pub := amqp.Publishing{MessageId: "<id>"}
			PackTrace(&pub, "<trace>")

			Expect(pub.CorrelationId).To(Equal("<trace>"))
		})

		It("does not set the correlation ID if the trace ID the same as the message ID", func() {
			pub := amqp.Publishing{MessageId: "<id>"}
			PackTrace(&pub, "<id>")

			Expect(pub.CorrelationId).To(Equal(""))
		})
	})

	Describe("UnpackTrace", func() {
		It("returns a the trace ID based on the message ID", func() {
			del := amqp.Delivery{MessageId: "<id>"}
			id := UnpackTrace(&del)

			Expect(id).To(Equal("<id>"))
		})

		It("returns a the trace ID based on the correlation ID", func() {
			del := amqp.Delivery{
				MessageId:     "<id>",
				CorrelationId: "<trace>",
			}
			id := UnpackTrace(&del)

			Expect(id).To(Equal("<trace>"))
		})
	})
})
