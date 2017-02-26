package notifyamqp

import (
	"context"

	"github.com/over-pass/overpass-go/src/internal/amqputil"
	"github.com/over-pass/overpass-go/src/internal/notify"
	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/streadway/amqp"
)

type notifier struct {
	channels amqputil.ChannelPool
	logger   overpass.Logger
}

// newNotifier creates, initializes and returns a new notifier.
func newNotifier(
	channels amqputil.ChannelPool,
	logger overpass.Logger,
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
) (string, error) {
	msg := amqp.Publishing{
		MessageId: msgID.String(),
		Type:      notificationType,
		Body:      payload.Bytes(),
	}
	corrID := amqputil.PutCorrelationID(ctx, &msg)

	if err := n.send(unicastExchange, target.String(), msg); err != nil {
		return corrID, err
	}

	return corrID, nil
}

func (n *notifier) NotifyMulticast(
	ctx context.Context,
	msgID overpass.MessageID,
	constraint overpass.Constraint,
	notificationType string,
	payload *overpass.Payload,
) (string, error) {
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
		return corrID, err
	}

	return corrID, nil
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
