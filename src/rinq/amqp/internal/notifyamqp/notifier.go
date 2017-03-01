package notifyamqp

import (
	"context"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/amqp/internal/amqputil"
	"github.com/rinq/rinq-go/src/rinq/internal/notify"
	"github.com/streadway/amqp"
)

type notifier struct {
	channels amqputil.ChannelPool
	logger   rinq.Logger
}

// newNotifier creates, initializes and returns a new notifier.
func newNotifier(
	channels amqputil.ChannelPool,
	logger rinq.Logger,
) notify.Notifier {
	return &notifier{
		channels: channels,
		logger:   logger,
	}
}

func (n *notifier) NotifyUnicast(
	ctx context.Context,
	msgID rinq.MessageID,
	target rinq.SessionID,
	notificationType string,
	payload *rinq.Payload,
) (string, error) {
	msg := amqp.Publishing{
		MessageId: msgID.String(),
		Type:      notificationType,
		Body:      payload.Bytes(),
	}
	traceID := amqputil.PackTrace(ctx, &msg)

	if err := n.send(unicastExchange, target.String(), msg); err != nil {
		return traceID, err
	}

	return traceID, nil
}

func (n *notifier) NotifyMulticast(
	ctx context.Context,
	msgID rinq.MessageID,
	constraint rinq.Constraint,
	notificationType string,
	payload *rinq.Payload,
) (string, error) {
	msg := amqp.Publishing{
		MessageId: msgID.String(),
		Type:      notificationType,
		Headers:   amqp.Table{},
		Body:      payload.Bytes(),
	}
	traceID := amqputil.PackTrace(ctx, &msg)

	for key, value := range constraint {
		msg.Headers[key] = value
	}

	if err := n.send(multicastExchange, "", msg); err != nil {
		return traceID, err
	}

	return traceID, nil
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
