package amqputil_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq/trace"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/amqputil"
	"github.com/streadway/amqp"
)

var _ = Describe("Trace", func() {
	Describe("PackTrace", func() {
		It("sets the correlation ID", func() {
			pub := amqp.Publishing{MessageId: "<id>"}
			amqputil.PackTrace(&pub, "<trace>")

			Expect(pub.CorrelationId).To(Equal("<trace>"))
		})

		It("does not set the correlation ID if the trace ID the same as the message ID", func() {
			pub := amqp.Publishing{MessageId: "<id>"}
			amqputil.PackTrace(&pub, "<id>")

			Expect(pub.CorrelationId).To(Equal(""))
		})
	})

	Describe("UnpackTrace", func() {
		It("returns a context with the trace ID based on the message ID", func() {
			del := amqp.Delivery{MessageId: "<id>"}
			ctx := amqputil.UnpackTrace(context.Background(), &del)

			Expect(trace.Get(ctx)).To(Equal("<id>"))
		})

		It("returns a context with the trace ID based on the correlation ID", func() {
			del := amqp.Delivery{
				MessageId:     "<id>",
				CorrelationId: "<trace>",
			}
			ctx := amqputil.UnpackTrace(context.Background(), &del)

			Expect(trace.Get(ctx)).To(Equal("<trace>"))
		})
	})
})
