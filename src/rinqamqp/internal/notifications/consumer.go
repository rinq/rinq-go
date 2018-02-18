package notifications

import (
	"github.com/jmalloc/twelf/src/twelf"
	"github.com/rinq/rinq-go/src/internal/service"
	"github.com/rinq/rinq-go/src/internal/transport"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/amqpx"
	"github.com/streadway/amqp"
)

// Consumer is an implementation of transport.Consumer that decodes AMQP
// messages to produce notifications.
type Consumer struct {
	service.Service
	sm *service.StateMachine

	peerID   ident.PeerID
	channels amqpx.ChannelPool
	decoder  *Decoder
	logger   twelf.Logger
	queue    chan transport.InboundNotification
}

// NewConsumer returns a new AMQP-based notification consumer.
func NewConsumer(
	peerID ident.PeerID,
	channels amqpx.ChannelPool,
	decoder *Decoder,
	logger twelf.Logger,
	preFetch uint,
) *Consumer {
	if preFetch == 0 {
		panic("consumer must use a pre-fetch limit")
	}

	c := &Consumer{
		peerID:   peerID,
		channels: channels,
		decoder:  decoder,
		logger:   logger,
		queue:    make(chan transport.InboundNotification, preFetch),
	}

	c.sm = service.NewStateMachine(c.run, c.finalize)
	c.Service = c.sm

	go c.sm.Run()

	return c
}

// Queue returns a channel on which inbound notifications are delivered.
func (c *Consumer) Queue() <-chan transport.InboundNotification {
	return c.queue
}

// Run consumes notifications until done is closed.
func (c *Consumer) run() (service.State, error) {
	defer close(c.queue)

	ch, err := c.channels.GetQOS(uint(cap(c.queue)))
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	closed := make(chan *amqp.Error)
	ch.NotifyClose(closed)

	queue := notifyQueue(c.peerID)
	deliveries, err := ch.Consume(
		queue,
		queue, // use queue name as consumer tag
		false, // autoAck
		true,  // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return nil, err
	}

	logConsumerStart(c.logger, c.peerID, cap(c.queue))

	for {
		select {
		case msg := <-deliveries:
			c.dispatch(&msg)
		case <-c.sm.Graceful:
			return nil, nil
		case <-c.sm.Forceful:
			return nil, nil
		case err := <-closed:
			return nil, err
		}
	}
}

func (c *Consumer) finalize(err error) error {
	logConsumerStop(c.logger, c.peerID, err)
	return err
}

func (c *Consumer) dispatch(msg *amqp.Delivery) {
	n, err := c.decoder.Unmarshal(msg)
	if err != nil {
		logConsumerIgnoredMessage(c.logger, c.peerID, msg, err)
		_ = msg.Reject(false) // false = don't requeue, error handled by AMQP close notification
	}

	// The pre-fetch count is based on c.queue's buffer size, and therefore
	// this cannot block.
	c.queue <- transport.InboundNotification{
		Notification: n,
		Done: func() {
			_ = msg.Ack(false) // false = single message, error handled by AMQP close notification
		},
	}
}
