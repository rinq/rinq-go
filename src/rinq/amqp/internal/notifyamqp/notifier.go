package notifyamqp

import (
	"context"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/amqp/internal/amqputil"
	"github.com/rinq/rinq-go/src/rinq/ident"
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
	msgID ident.MessageID,
	target ident.SessionID,
	notificationType string,
	payload *rinq.Payload,
) (traceID string, err error) {
	msg := amqp.Publishing{
		MessageId: msgID.String(),
		Type:      notificationType,
		Body:      payload.Bytes(),
	}
	traceID = amqputil.PackTrace(ctx, &msg)
	err = n.send(unicastExchange, target.String(), msg)
	return
}

func (n *notifier) NotifyMulticast(
	ctx context.Context,
	msgID ident.MessageID,
	constraint rinq.Constraint,
	notificationType string,
	payload *rinq.Payload,
) (traceID string, err error) {
	msg := amqp.Publishing{
		MessageId: msgID.String(),
		Type:      notificationType,
		Body:      payload.Bytes(),
	}

	packConstraint(&msg, constraint)

	traceID = amqputil.PackTrace(ctx, &msg)
	err = n.send(multicastExchange, "", msg)
	return
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
