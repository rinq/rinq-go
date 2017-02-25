package notifyamqp

import (
	"context"
	"fmt"
	"sync"

	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/over-pass/overpass-go/src/internals/localsession"
	"github.com/over-pass/overpass-go/src/internals/notify"
	"github.com/over-pass/overpass-go/src/internals/revision"
	"github.com/over-pass/overpass-go/src/internals/service"
	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/streadway/amqp"
)

type listener struct {
	service.Service
	closer *service.Closer

	peerID    overpass.PeerID
	preFetch  int
	sessions  localsession.Store
	revisions revision.Store
	logger    overpass.Logger
	waiter    sync.WaitGroup

	mutex    sync.RWMutex
	channel  *amqp.Channel
	handlers map[overpass.SessionID]overpass.NotificationHandler
}

// newListener creates, starts and returns a new listener.
func newListener(
	peerID overpass.PeerID,
	preFetch int,
	sessions localsession.Store,
	revisions revision.Store,
	channel *amqp.Channel,
	logger overpass.Logger,
) (notify.Listener, error) {
	svc, closer := service.NewImpl()

	l := &listener{
		Service: svc,
		closer:  closer,

		peerID:    peerID,
		preFetch:  preFetch,
		sessions:  sessions,
		revisions: revisions,
		logger:    logger,
		channel:   channel,
		handlers:  map[overpass.SessionID]overpass.NotificationHandler{},
	}

	if err := l.initialize(); err != nil {
		return nil, err
	}

	return l, nil
}

func (l *listener) Listen(id overpass.SessionID, handler overpass.NotificationHandler) (bool, error) {
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
			return false, err
		}

		if err := l.channel.QueueBind(
			queue,
			"",
			multicastExchange,
			false, // noWait
			nil,   // args
		); err != nil {
			return false, err
		}
	}

	_, exists := l.handlers[id]
	l.handlers[id] = handler

	return !exists, nil
}

func (l *listener) Unlisten(id overpass.SessionID) (bool, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	_, exists := l.handlers[id]
	delete(l.handlers, id)

	if len(l.handlers) == 0 {
		queue := notifyQueue(l.peerID)

		if err := l.channel.QueueUnbind(
			queue,
			id.Peer.String()+".*",
			unicastExchange,
			nil, // args
		); err != nil {
			return false, err
		}

		if err := l.channel.QueueUnbind(
			queue,
			"",
			multicastExchange,
			nil, // args
		); err != nil {
			return false, err
		}
	}

	return exists, nil
}

func (l *listener) initialize() error {
	queue := notifyQueue(l.peerID)

	if err := l.channel.Qos(l.preFetch, 0, true); err != nil {
		return err
	}

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
		false, // autoAck
		true,  // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return err
	}

	go l.monitor()
	go l.dispatchEach(messages)

	return nil
}

func (l *listener) monitor() {
	logListenerStart(l.logger, l.peerID, l.preFetch)

	var err error

	select {
	case err = <-l.channel.NotifyClose(make(chan *amqp.Error)):
	case <-l.closer.Stop():
		l.channel.Close()
		if l.closer.IsGraceful() {
			logListenerStopping(l.logger, l.peerID)
			l.waiter.Wait()
		}
	}

	l.closer.Close(err)

	logListenerStop(l.logger, l.peerID, err)
}

func (l *listener) dispatchEach(messages <-chan amqp.Delivery) {
	for msg := range messages {
		l.waiter.Add(1)
		go l.dispatch(msg)
	}
}

func (l *listener) dispatch(msg amqp.Delivery) {
	defer l.waiter.Done()

	msgID, err := overpass.ParseMessageID(msg.MessageId)
	if err != nil {
		if l.logger.IsDebug() {
			l.logger.Log(
				"%s notification listener ignored AMQP message, '%s' is not a valid message ID",
				l.peerID.ShortString(),
				msg.MessageId,
			)
		}
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

	if err != nil && l.logger.IsDebug() {
		l.logger.Log(
			"%s notification listener ignored AMQP message %s, %s",
			l.peerID.ShortString(),
			msgID.ShortString(),
			err,
		)
	}

	msg.Ack(false)
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
		amqputil.WithCorrelationID(context.Background(), &msg),
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

	ctx := amqputil.WithCorrelationID(context.Background(), &msg)
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

		if err != nil && l.logger.IsDebug() {
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

	l.mutex.RLock()
	handler := l.handlers[sess.ID()]
	l.mutex.RUnlock()

	if handler == nil {
		return nil
	}

	handler(ctx, sess, n)

	// set the payload pointer to null now that it's the handler's
	// responsibility. calling close on a nil payload pointer is safe.
	n.Payload = nil

	return nil
}
