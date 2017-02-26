package amqputil_test

import (
	"context"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/over-pass/overpass-go/src/overpassamqp/internal/amqputil"
	"github.com/streadway/amqp"
)

var _ = Describe("Deadline", func() {
	Describe("PackDeadline", func() {
		It("sets the expiration", func() {
			now := time.Now()
			deadline := now.Add(10 * time.Second)

			ctx, cancel := context.WithDeadline(context.Background(), deadline)
			defer cancel()

			msg := amqp.Publishing{}
			hasDeadline, err := amqputil.PackDeadline(ctx, &msg)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(hasDeadline).To(BeTrue())

			expiration, err := strconv.ParseUint(msg.Expiration, 10, 64)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(expiration).Should(BeNumerically("~", (10*time.Second)/time.Millisecond, 10))
			Expect(msg.Headers["dl"].(int64)).To(Equal(deadline.UnixNano() / int64(time.Millisecond)))
		})

		It("returns an error if the deadline has already passed", func() {
			ctx, cancel := context.WithTimeout(context.Background(), -1)
			defer cancel()

			msg := amqp.Publishing{}
			hasDeadline, err := amqputil.PackDeadline(ctx, &msg)

			Expect(err).To(Equal(ctx.Err()))
			Expect(hasDeadline).To(BeTrue())
		})

		It("does nothing if the context has no deadline", func() {
			msg := amqp.Publishing{}
			hasDeadline, err := amqputil.PackDeadline(context.Background(), &msg)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(hasDeadline).To(BeFalse())

			Expect(msg.Expiration).To(Equal(""))
			Expect(msg.Timestamp.IsZero()).To(BeTrue())
		})
	})

	Describe("UnpackDeadline", func() {
		It("returns a context with the deadline from the message", func() {
			expected := time.Now()

			msg := amqp.Delivery{
				Headers:    amqp.Table{"dl": expected.UnixNano() / int64(time.Millisecond)},
				Expiration: "0",
			}

			ctx, cancel := amqputil.UnpackDeadline(context.Background(), &msg)
			defer cancel()

			deadline, ok := ctx.Deadline()

			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally("~", expected, time.Millisecond)) // within one milli
		})

		It("does not add a deadline if there is no deadline in the message", func() {
			msg := amqp.Delivery{
				Expiration: "1000",
			}

			ctx, cancel := amqputil.UnpackDeadline(context.Background(), &msg)
			defer cancel()

			_, ok := ctx.Deadline()

			Expect(ok).To(BeFalse())
		})
	})
})
