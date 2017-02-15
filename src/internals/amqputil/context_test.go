package amqputil_test

import (
	"context"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/streadway/amqp"
)

var _ = Describe("Context", func() {
	Describe("PutCorrelationID", func() {
		It("sets the correlation ID", func() {
			del := amqp.Delivery{MessageId: "<id>"}
			ctx := amqputil.WithCorrelationID(context.Background(), del)

			pub := amqp.Publishing{}
			result := amqputil.PutCorrelationID(ctx, &pub)

			Expect(pub.CorrelationId).To(Equal(del.MessageId))
			Expect(result).To(Equal(del.MessageId))
		})

		It("does not set the correlation ID if it's the same as a the message ID", func() {
			del := amqp.Delivery{MessageId: "<id>"}
			ctx := amqputil.WithCorrelationID(context.Background(), del)

			pub := amqp.Publishing{MessageId: "<id>"}
			result := amqputil.PutCorrelationID(ctx, &pub)

			Expect(pub.CorrelationId).To(Equal(""))
			Expect(result).To(Equal(del.MessageId))
		})
	})

	Describe("WithCorrelationID and GetCorrelationID", func() {
		It("transports the correlation ID in the context", func() {
			msg := amqp.Delivery{MessageId: "<id>"}
			ctx := amqputil.WithCorrelationID(context.Background(), msg)

			Expect(amqputil.GetCorrelationID(ctx)).To(Equal("<id>"))
		})
	})

	Describe("PutExpiration", func() {
		It("sets the timestamp and expiration", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			msg := amqp.Publishing{}
			calledAt := time.Now()
			hasDeadline, err := amqputil.PutExpiration(ctx, &msg)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(hasDeadline).To(BeTrue())

			expiration, err := strconv.ParseUint(msg.Expiration, 10, 64)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(expiration).Should(BeNumerically("~", (10*time.Second)/time.Millisecond, 100))
			Expect(msg.Timestamp).Should(BeTemporally("~", calledAt, 10*time.Millisecond))
		})

		It("returns an error if the deadline has already passed", func() {
			ctx, cancel := context.WithTimeout(context.Background(), -1)
			defer cancel()

			msg := amqp.Publishing{}
			hasDeadline, err := amqputil.PutExpiration(ctx, &msg)

			Expect(err).To(Equal(ctx.Err()))
			Expect(hasDeadline).To(BeTrue())
		})

		It("does nothing if the context has no deadline", func() {
			msg := amqp.Publishing{}
			hasDeadline, err := amqputil.PutExpiration(context.Background(), &msg)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(hasDeadline).To(BeFalse())

			Expect(msg.Expiration).To(Equal(""))
			Expect(msg.Timestamp.IsZero()).To(BeTrue())
		})
	})

	Describe("WithExpiration", func() {
		It("it adds the deadline from the message", func() {
			msg := amqp.Delivery{
				Timestamp:  time.Now(),
				Expiration: "10000",
			}

			ctx, cancel := amqputil.WithExpiration(context.Background(), msg)
			defer cancel()

			deadline, ok := ctx.Deadline()

			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally("==", msg.Timestamp.Add(1000*time.Millisecond)))
		})

		It("does not add a deadline if there is no timestamp", func() {
			msg := amqp.Delivery{}

			ctx, cancel := amqputil.WithExpiration(context.Background(), msg)
			defer cancel()

			_, ok := ctx.Deadline()

			Expect(ok).To(BeFalse())
		})

		It("does not add a deadline if there the expiration is not an integer", func() {
			msg := amqp.Delivery{
				Timestamp:  time.Now(),
				Expiration: "<string>",
			}

			ctx, cancel := amqputil.WithExpiration(context.Background(), msg)
			defer cancel()

			_, ok := ctx.Deadline()

			Expect(ok).To(BeFalse())
		})
	})
})
