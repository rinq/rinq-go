package amqputil_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/over-pass/overpass-go/src/overpassamqp/internal/amqputil"
	"github.com/streadway/amqp"
)

var _ = Describe("Trace", func() {
	Describe("PackTrace and UnpackTrace", func() {
		It("sets the correlation ID", func() {
			del := amqp.Delivery{MessageId: "<id>"}
			ctx := amqputil.UnpackTrace(context.Background(), &del)

			pub := amqp.Publishing{}
			result := amqputil.PackTrace(ctx, &pub)

			Expect(pub.CorrelationId).To(Equal(del.MessageId))
			Expect(result).To(Equal(del.MessageId))
		})

		It("does not set the correlation ID if it's the same as the message ID", func() {
			del := amqp.Delivery{MessageId: "<id>"}
			ctx := amqputil.UnpackTrace(context.Background(), &del)

			pub := amqp.Publishing{MessageId: "<id>"}
			result := amqputil.PackTrace(ctx, &pub)

			Expect(pub.CorrelationId).To(Equal(""))
			Expect(result).To(Equal(del.MessageId))
		})
	})
})
