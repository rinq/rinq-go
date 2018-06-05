package amqpx_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/amqptest"
	. "github.com/rinq/rinq-go/src/rinqamqp/internal/amqpx"
	"github.com/streadway/amqp"
)

var _ = Describe("ChannelPool", func() {
	var (
		broker   *amqp.Connection
		channels ChannelPool
	)

	BeforeSuite(func() {
		broker = amqptest.Connect()
		channels = NewChannelPool(broker, 3)
	})

	AfterSuite(func() {
		broker.Close()
	})

	Describe("Get", func() {
		It("returns a new channel", func() {
			c, err := channels.Get()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(c).ToNot(BeNil())
			c.Close()
		})

		It("reuses existing channels", func() {
			c1, err := channels.Get()
			Expect(err).ShouldNot(HaveOccurred())

			channels.Put(c1)

			c2, err := channels.Get()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(c2).To(BeIdenticalTo(c1))

			c2.Close()
		})
	})

	Describe("GetQOS", func() {
		It("returns a new channel", func() {
			c, err := channels.GetQOS(1)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(c).ToNot(BeNil())

			c.Close()
		})

		It("sets the pre-fetch count", func() {
			c, err := channels.GetQOS(1)
			Expect(err).ShouldNot(HaveOccurred())
			defer c.Close()

			q, err := c.QueueDeclare(
				"",
				false, // durable
				false, // autoDelete
				true,  // exclusive
				false, // noWait
				nil,   // args
			)
			Expect(err).ShouldNot(HaveOccurred())

			msgs, err := c.Consume(
				q.Name,
				q.Name, // consumer name
				false,  // autoAck
				true,   // exclusive
				false,  // noLocal
				false,  // noWait
				nil,    // args
			)
			Expect(err).ShouldNot(HaveOccurred())

			err = c.Publish("", q.Name, false, false, amqp.Publishing{})
			Expect(err).ShouldNot(HaveOccurred())

			err = c.Publish("", q.Name, false, false, amqp.Publishing{})
			Expect(err).ShouldNot(HaveOccurred())

			// canceling the consumer does NOT affect already delivered messages
			err = c.Cancel(q.Name, false)
			Expect(err).ShouldNot(HaveOccurred())

			// read first message
			select {
			case _, ok := <-msgs:
				Expect(ok).To(BeTrue())
			case <-time.After(time.Second):
				panic("no message delivered")
			}

			// wait for close
			select {
			case _, ok := <-msgs:
				// we shouldn't get a second message here
				Expect(ok).To(BeFalse())
			case <-time.After(time.Second):
				panic("channel not closed")
			}
		})
	})
})
