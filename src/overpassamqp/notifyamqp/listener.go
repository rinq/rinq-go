package notifyamqp

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/over-pass/overpass-go/src/internals/deferutil"
	"github.com/over-pass/overpass-go/src/internals/localsession"
	"github.com/over-pass/overpass-go/src/internals/notify"
	"github.com/over-pass/overpass-go/src/internals/revision"
	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/streadway/amqp"
)

type listener struct {
	peerID    overpass.PeerID
	sessions  localsession.Store
	revisions revision.Store
	logger    overpass.Logger

	mutex    sync.RWMutex
	channel  *amqp.Channel
	handlers map[overpass.SessionID]overpass.NotificationHandler

	done chan struct{}
	err  atomic.Value
}

// newListener creates, starts and returns a new listener.
func newListener(
	peerID overpass.PeerID,
	sessions localsession.Store,
	revisions revision.Store,
	channel *amqp.Channel,
	logger overpass.Logger,
) (notify.Listener, error) {
	l := &listener{
		peerID:    peerID,
		sessions:  sessions,
		revisions: revisions,
		logger:    logger,
		channel:   channel,
		handlers:  map[overpass.SessionID]overpass.NotificationHandler{},
		done:      make(chan struct{}),
	}

	if err := l.initialize(); err != nil {
		return nil, err
	}

	return l, nil
}

func (l *listener) Listen(id overpass.SessionID, handler overpass.NotificationHandler) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if len(l.handlers) == 0 {
		queue := notifyQueue(l.peerID)

		if err := l.channel.QueueBind(
			queue,
			id.Peer.String()+".*",
			unicastExchange,
			false, // noWait
			nil,   // args
		); err != nil {
			return err
		}

		if err := l.channel.QueueBind(
			queue,
			"",
			multicastExchange,
			false, // noWait
			nil,   // args
		); err != nil {
			return err
		}
	}

	l.handlers[id] = handler
	return nil
}

func (l *listener) Unlisten(id overpass.SessionID) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	delete(l.handlers, id)

	if len(l.handlers) == 0 {
		queue := notifyQueue(l.peerID)

		if err := l.channel.QueueUnbind(
			queue,
			id.Peer.String()+".*",
			unicastExchange,
			nil, // args
		); err != nil {
			return err
		}

		if err := l.channel.QueueUnbind(
			queue,
			"",
			multicastExchange,
			nil, // args
		); err != nil {
			return err
		}
	}

	return nil
}

func (l *listener) Done() <-chan struct{} {
	return l.done
}

func (l *listener) Error() error {
	if err, ok := l.err.Load().(error); ok {
		return err
	}

	return nil
}

func (l *listener) initialize() error {
	queue := notifyQueue(l.peerID)

	if _, err := l.channel.QueueDeclare(
		queue,
		false, // durable
		false, // autoDelete
		true,  // exclusive,
		false, // noWait
		nil,   // args
	); err != nil {
		return err
	}

	messages, err := l.channel.Consume(
		queue,
		queue, // use queue name as consumer tag
		true,  // autoAck
		true,  // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return err
	}

	go l.consume(messages)

	return nil
}

func (l *listener) consume(messages <-chan amqp.Delivery) {
	done := l.channel.NotifyClose(make(chan *amqp.Error))

	for msg := range messages {
		l.dispatch(msg)
	}

	if amqpErr := <-done; amqpErr != nil {
		// we can't just return err when it's nil, because it will be a nil
		// *amqp.Error, as opposed to a nil "error" interface.
		l.close(amqpErr)
	} else {
		l.close(nil)
	}
}

func (l *listener) dispatch(msg amqp.Delivery) {
	msgID, err := overpass.ParseMessageID(msg.MessageId)
	if err != nil {
		l.logger.Log(
			"%s ignored AMQP message, '%s' is not a valid message ID",
			l.peerID.ShortString(),
			msg.MessageId,
		)
		return
	}

	switch msg.Exchange {
	case unicastExchange:
		err = l.handleUnicast(msgID, msg)
	case multicastExchange:
		err = l.handleMulticast(msgID, msg)
	default:
		err = fmt.Errorf("delivery via '%s' exchange is not expected", msg.Exchange)
	}

	if err != nil {
		l.logger.Log(
			"%s ignored AMQP message %s, %s",
			l.peerID.ShortString(),
			msgID.ShortString(),
			err,
		)
	}
}

func (l *listener) handleUnicast(msgID overpass.MessageID, msg amqp.Delivery) error {
	sessID, err := overpass.ParseSessionID(msg.RoutingKey)
	if err != nil {
		return err
	}

	sess, _, ok := l.sessions.Get(sessID)
	if !ok {
		return nil
	} else if err != nil {
		return err
	}

	source, err := l.revisions.GetRevision(msgID.Session)
	if err != nil {
		return err
	}

	return l.handle(
		amqputil.WithCorrelationID(context.Background(), msg),
		msgID,
		sess,
		overpass.Notification{
			Source:  source,
			Type:    msg.Type,
			Payload: overpass.NewPayloadFromBytes(msg.Body),
		},
	)
}

func (l *listener) handleMulticast(msgID overpass.MessageID, msg amqp.Delivery) error {
	constraint := overpass.Constraint{}
	for key, value := range msg.Headers {
		if v, ok := value.(string); ok {
			constraint[key] = v
		} else {
			return fmt.Errorf("constraint key %s contains non-string value", key)
		}
	}

	var sessions []overpass.Session

	l.sessions.Each(func(session overpass.Session, catalog localsession.Catalog) {
		_, attrs := catalog.Attrs()
		if attrs.MatchConstraint(constraint) {
			sessions = append(sessions, session)
		}
	})

	if len(sessions) == 0 {
		return nil
	}

	source, err := l.revisions.GetRevision(msgID.Session)
	if err != nil {
		return err
	}

	ctx := amqputil.WithCorrelationID(context.Background(), msg)
	payload := overpass.NewPayloadFromBytes(msg.Body)
	defer payload.Close()

	for _, sess := range sessions {
		err = l.handle(
			ctx,
			msgID,
			sess,
			overpass.Notification{
				Source:      source,
				Type:        msg.Type,
				Payload:     payload.Clone(),
				IsMulticast: true,
				Constraint:  constraint,
			},
		)

		if err != nil {
			l.logger.Log(
				"%s ignored notification %s for %s, %s",
				l.peerID.ShortString(),
				msgID.ShortString(),
				sess.ID().ShortString(),
				err,
			)
		}
	}

	return nil
}

func (l *listener) handle(
	ctx context.Context,
	msgID overpass.MessageID,
	sess overpass.Session,
	n overpass.Notification,
) error {
	// we want to close the payload if the handler is never called
	defer n.Payload.Close()

	var handler overpass.NotificationHandler
	deferutil.RWith(&l.mutex, func() {
		handler = l.handlers[sess.ID()]
	})

	if handler == nil {
		return nil
	}

	rev, err := sess.CurrentRevision()
	if overpass.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	l.logger.Log(
		"%s received '%s' notification from %s (%d bytes) [%s]",
		rev.Ref().ShortString(),
		n.Type,
		n.Source.Ref().ShortString(),
		n.Payload.Len(),
		amqputil.GetCorrelationID(ctx),
	)

	handler(ctx, sess, n)

	// set the payload pointer to null now that it's the handler's
	// responsibility. calling close on a nil payload pointer is safe.
	n.Payload = nil

	return nil
}

func (l *listener) close(err error) {
	if err != nil {
		l.err.Store(err)
	}
	close(l.done)
	l.channel.Close() // TODO lock mutes
}
