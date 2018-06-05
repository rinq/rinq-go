package marshaling_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/amqptest"
	. "github.com/rinq/rinq-go/src/rinqamqp/internal/marshaling"
	"github.com/streadway/amqp"
)

var _ = Describe("PackTrace and UnpackTrace", func() {
	It("transmits the trace ID", func() {
		pub := &amqp.Publishing{MessageId: "<id>"}
		PackTrace(pub, "<trace>")

		del := amqptest.PublishingToDelivery(pub)
		trace := UnpackTrace(del)

		Expect(trace).To(Equal("<trace>"))
	})

	It("transmits the trace ID when it's the same as the message ID", func() {
		pub := &amqp.Publishing{MessageId: "<id>"}
		PackTrace(pub, "<id>")

		del := amqptest.PublishingToDelivery(pub)
		trace := UnpackTrace(del)

		Expect(trace).To(Equal("<id>"))
	})
})
