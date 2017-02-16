package notifyamqp

import (
	"context"
	"log"

	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/over-pass/overpass-go/src/internals/notify"
	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/streadway/amqp"
)

type notifier struct {
	channels amqputil.ChannelPool
	logger   *log.Logger
}

// newNotifier creates, initializes and returns a new notifier.
func newNotifier(
	channels amqputil.ChannelPool,
	logger *log.Logger,
) notify.Notifier {
	return &notifier{
		channels: channels,
		logger:   logger,
	}
}

func (n *notifier) NotifyUnicast(
	ctx context.Context,
	msgID overpass.MessageID,
	target overpass.SessionID,
	notificationType string,
	payload *overpass.Payload,
) error {
	msg := amqp.Publishing{
		MessageId: msgID.String(),
		Type:      notificationType,
		Body:      payload.Bytes(),
	}
	corrID := amqputil.PutCorrelationID(ctx, &msg)

	if err := n.send(unicastExchange, target.String(), msg); err != nil {
		return err
	}

	n.logger.Printf(
		"%s sent '%s' notification to %s (%d bytes) [%s]",
		msgID.ShortString(),
		notificationType,
		target.ShortString(),
		payload.Len(),
		corrID,
	)

	return nil
}

func (n *notifier) NotifyMulticast(
	ctx context.Context,
	msgID overpass.MessageID,
	constraint overpass.Constraint,
	notificationType string,
	payload *overpass.Payload,
) error {
	msg := amqp.Publishing{
		MessageId: msgID.String(),
		Type:      notificationType,
		Headers:   amqp.Table{},
		Body:      payload.Bytes(),
	}
	corrID := amqputil.PutCorrelationID(ctx, &msg)

	for key, value := range constraint {
		msg.Headers[key] = value
	}

	if err := n.send(multicastExchange, "", msg); err != nil {
		return err
	}

	n.logger.Printf(
		"%s sent '%s' notification to {%s} (%d bytes) [%s]",
		msgID.ShortString(),
		notificationType,
		constraint,
		payload.Len(),
		corrID,
	)

	return nil
}

func (n *notifier) send(exchange, key string, msg amqp.Publishing) error {
	channel, err := n.channels.Get()
	if err != nil {
		return err
	}
	defer n.channels.Put(channel)

	return channel.Publish(
		exchange,
		key,
		false, // mandatory
		false, // immediate
		msg,
	)
}
