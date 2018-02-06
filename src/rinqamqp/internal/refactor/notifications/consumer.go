package notifications

import (
	"context"
	"errors"

	"github.com/rinq/rinq-go/src/internal/notifications"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/refactor/amqpx"
	"github.com/streadway/amqp"
	"github.com/uber-go/multierr"
)

// Consumer consumes AMQP notification messages and sends them to a channel
// compatible with notifications.Source.
type Consumer struct {
	Context       context.Context
	Broker        *amqp.Connection
	PeerID        ident.PeerID
	Notifications chan<- notifications.Inbound
	Decoder       *Decoder
	Logger        rinq.Logger

	channel    *amqp.Channel
	closed     chan *amqp.Error
	deliveries <-chan amqp.Delivery
}

// Run processes incoming notifications until c.Context is canceled.
// c.Notifications is closed when Run returns.
func (c *Consumer) Run() error {
	defer close(c.Notifications)

	if err := c.initChannel(); err != nil {
		return err
	}

	if err := c.startConsumer(); err != nil {
		return err
	}

	logConsumerStart(c.Logger, c.PeerID, cap(c.Notifications))

	var err error
	for err != nil {
		select {
		case msg, ok := <-c.deliveries:
			if ok {
				c.process(&msg)
			} else {
				// if the consumer channel is closed before we call
				// stopConsumer() then there must be an AMQP error.
				c.deliveries = nil
			}

		case err = <-c.closed:
			if err == nil {
				err = errors.New("AMQP channel closed unexpectedly")
			}

		case <-c.Context.Done():
			err = c.Context.Err()

			// canceling the context is the standard way to stop the consumer, and
			// hence is not an error.
			if err == context.Canceled {
				err = nil
			}

			if e := c.stopConsumer(); e != nil {
				err = multierr.Append(err, e)
			}
		}
	}

	logConsumerStop(c.Logger, c.PeerID, err)

	return err
}

// initChannel opens a new AMQP channel and configures if for consuming.
func (c *Consumer) initChannel() error {
	preFetch := cap(c.Notifications)
	if preFetch == 0 {
		// preFetch of zero means unlimited. instead, if the target channel is
		// unbuffered we only accept one AMQP message at a time.
		preFetch = 1
	}

	var err error
	c.channel, err = amqpx.ChannelWithPreFetch(c.Broker, preFetch)
	if err != nil {
		return err
	}

	c.closed = make(chan *amqp.Error, 1) // buffer of 1 as we may never read this
	c.channel.NotifyClose(c.closed)

	return nil
}

// process decodes msg and sends the resulting notification to c.Notifications.
func (c *Consumer) process(msg *amqp.Delivery) {
	n, err := c.Decoder.Unmarshal(msg)
	if err != nil {
		_ = msg.Reject(false) // false = don't requeue
		logIgnoredMessage(c.Logger, c.PeerID, msg, err)
		return
	}

	// this cannot block, as the AMQP channel's pre-fetch count is set to
	// the Go channel's capacity
	c.Notifications <- notifications.Inbound{
		Notification: n,
		Ack: func() {
			_ = msg.Ack(false) // false = single message
		},
	}
}

// startConsumer starts the AMQP consumer that receives notification messages.
func (c *Consumer) startConsumer() error {
	queue := notifyQueue(c.PeerID)

	var err error
	c.deliveries, err = c.channel.Consume(
		queue,
		queue, // use queue name as consumer tag
		false, // autoAck
		true,  // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)

	return err
}

// stopConsumer cancels the AMQP consumer and rejects any new messages that are
// received while the cancelation is in progress.
func (c *Consumer) stopConsumer() error {
	if err := c.channel.Cancel(
		notifyQueue(c.PeerID), // use queue name as consumer tag
		false, // noWait
	); err != nil {
		return err
	}

	// reject any messages that have already been delivered.
	// s.deliveries is eventually closed due to the call to Cancel() above.
	for msg := range c.deliveries {
		if err := msg.Reject(false); err != nil { // false = don't requeue
			return err
		}
	}

	return nil
}
