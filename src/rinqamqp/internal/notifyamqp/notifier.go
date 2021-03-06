package notifyamqp

import (
	"context"

	"github.com/jmalloc/twelf/src/twelf"
	"github.com/rinq/rinq-go/src/internal/notify"
	"github.com/rinq/rinq-go/src/internal/service"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/constraint"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/amqputil"
	"github.com/streadway/amqp"
)

type notifier struct {
	service.Service
	sm *service.StateMachine

	peerID   ident.PeerID
	channels amqputil.ChannelPool
	logger   twelf.Logger
}

// newNotifier creates, initializes and returns a new notifier.
func newNotifier(
	peerID ident.PeerID,
	channels amqputil.ChannelPool,
	logger twelf.Logger,
) notify.Notifier {
	n := &notifier{
		peerID:   peerID,
		channels: channels,
		logger:   logger,
	}

	n.sm = service.NewStateMachine(n.run, n.finalize)
	n.Service = n.sm

	go n.sm.Run()

	return n
}

func (n *notifier) NotifyUnicast(
	ctx context.Context,
	msgID ident.MessageID,
	traceID string,
	target ident.SessionID,
	ns string,
	notificationType string,
	payload *rinq.Payload,
) (err error) {
	msg := amqp.Publishing{
		MessageId: msgID.String(),
	}

	packCommonAttributes(&msg, traceID, ns, notificationType, payload)
	packTarget(&msg, target)

	err = amqputil.PackSpanContext(ctx, &msg)

	if err == nil {
		err = n.send(unicastExchange, unicastRoutingKey(ns, target.Peer), msg)
	}

	return
}

func (n *notifier) NotifyMulticast(
	ctx context.Context,
	msgID ident.MessageID,
	traceID string,
	con constraint.Constraint,
	ns string,
	notificationType string,
	payload *rinq.Payload,
) (err error) {
	msg := amqp.Publishing{
		MessageId: msgID.String(),
	}

	packCommonAttributes(&msg, traceID, ns, notificationType, payload)
	packConstraint(&msg, con)

	err = amqputil.PackSpanContext(ctx, &msg)

	if err == nil {
		err = n.send(multicastExchange, ns, msg)
	}

	return
}

func (n *notifier) send(exchange, key string, msg amqp.Publishing) error {
	select {
	case <-n.sm.Graceful:
		return context.Canceled
	case <-n.sm.Forceful:
		return context.Canceled
	default:
		// ready to publish
	}

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

func (n *notifier) run() (service.State, error) {
	logNotifierStart(n.logger, n.peerID)

	select {
	case <-n.sm.Graceful:
		return nil, nil

	case <-n.sm.Forceful:
		return nil, nil
	}
}

func (n *notifier) finalize(err error) error {
	logNotifierStop(n.logger, n.peerID, err)
	return err
}
