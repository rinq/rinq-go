package amqputil_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/streadway/amqp"
)

var _ = Describe("WithCorrelationID and GetCorrelationID", func() {
	It("transports the correlation ID in the context", func() {
		msg := amqp.Delivery{MessageId: "<id>"}
		ctx := amqputil.WithCorrelationID(context.Background(), msg)

		Expect(amqputil.GetCorrelationID(ctx)).To(Equal("<id>"))
	})
})
