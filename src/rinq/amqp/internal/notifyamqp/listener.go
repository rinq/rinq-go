package notifyamqp

import (
	"context"
	"fmt"
	"sync"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/amqp/internal/amqputil"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/localsession"
	"github.com/rinq/rinq-go/src/rinq/internal/notify"
	"github.com/rinq/rinq-go/src/rinq/internal/revision"
	"github.com/rinq/rinq-go/src/rinq/internal/service"
	"github.com/streadway/amqp"
)

type listener struct {
	service.Service
	sm *service.StateMachine

	peerID    ident.PeerID
	preFetch  int
	sessions  localsession.Store
	revisions revision.Store
	logger    rinq.Logger

	parentCtx context.Context // parent of all contexts passed to handlers
	cancelCtx func()          // cancels parentCtx when the server stops

	mutex    sync.RWMutex
	channel  *amqp.Channel // channel used for consuming
	handlers map[ident.SessionID]rinq.NotificationHandler

	deliveries <-chan amqp.Delivery // incoming notifications
	handled    chan struct{}        // signals a notification has been handled
	amqpClosed chan *amqp.Error

	// state-machine data
	pending uint // number of notifications currently being handled
}

// newListener creates, starts and returns a new listener.
func newListener(
	peerID ident.PeerID,
	preFetch int,
	sessions localsession.Store,
	revisions revision.Store,
	channel *amqp.Channel,
	logger rinq.Logger,
) (notify.Listener, error) {
	l := &listener{
		peerID:    peerID,
		preFetch:  preFetch,
		sessions:  sessions,
		revisions: revisions,
		logger:    logger,
		channel:   channel,

		handlers: map[ident.SessionID]rinq.NotificationHandler{},

		handled:    make(chan struct{}, preFetch),
		amqpClosed: make(chan *amqp.Error, 1),
	}

	l.sm = service.NewStateMachine(l.run, l.finalize)
	l.Service = l.sm

	if err := l.initialize(); err != nil {
		return nil, err
	}

	go l.sm.Run()

	return l, nil
}

func (l *listener) Listen(id ident.SessionID, handler rinq.NotificationHandler) (bool, error) {
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

func (l *listener) Unlisten(id ident.SessionID) (bool, error) {
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

// initialize prepares the AMQP channel
func (l *listener) initialize() error {
	l.channel.NotifyClose(l.amqpClosed)

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

	var err error
	l.deliveries, err = l.channel.Consume(
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

// run is the state entered when the service starts
func (l *listener) run() (service.State, error) {
	logListenerStart(l.logger, l.peerID, l.preFetch)

	l.parentCtx, l.cancelCtx = context.WithCancel(context.Background())

	for {
		select {
		case msg, ok := <-l.deliveries:
			if !ok {
				// sometimes the consumer channel is closed before the AMQP channel
				return nil, <-l.amqpClosed
			}
			l.pending++
			go l.dispatch(&msg)

		case <-l.handled:
			l.pending--

		case <-l.sm.Graceful:
			return l.graceful, nil

		case <-l.sm.Forceful:
			return l.forceful, nil

		case err := <-l.amqpClosed:
			return nil, err
		}
	}
}

// graceful is the state entered when a graceful stop is requested
func (l *listener) graceful() (service.State, error) {
	logListenerStopping(l.logger, l.peerID, l.pending)

	if err := l.closeChannel(); err != nil {
		return nil, err
	}

	for l.pending > 0 {
		select {
		case <-l.handled:
			l.pending--

		case <-l.sm.Forceful:
			return nil, nil
		}
	}

	return nil, nil
}

// forceful is the state entered when a stop is requested
func (l *listener) forceful() (service.State, error) {
	return nil, l.closeChannel()
}

// finalize is the state-machine finalizer, it is called immediately before the
// Done() channel is closed.
func (l *listener) finalize(err error) error {
	l.cancelCtx()
	logListenerStop(l.logger, l.peerID, err)
	return err
}

// dispatch validates an incoming notification and dispatches it the
// appropriate handler.
func (l *listener) dispatch(msg *amqp.Delivery) {
	defer func() {
		select {
		case l.handled <- struct{}{}:
		case <-l.sm.Forceful:
		}
	}()

	msgID, err := ident.ParseMessageID(msg.MessageId)
	if err != nil {
		msg.Reject(false)
		logInvalidMessageID(l.logger, l.peerID, msg.MessageId)
		return
	}

	// find the source session revision
	source, err := l.revisions.GetRevision(msgID.Ref)
	if err != nil {
		msg.Reject(false)
		logIgnoredMessage(l.logger, l.peerID, msgID, err)
		return
	}

	switch msg.Exchange {
	case unicastExchange:
		err = l.handleUnicast(msgID, msg, source)
	case multicastExchange:
		err = l.handleMulticast(msgID, msg, source)
	default:
		err = fmt.Errorf("delivery via '%s' exchange is not expected", msg.Exchange)
	}

	if err == nil {
		msg.Ack(false)
	} else {
		msg.Reject(false)
		logIgnoredMessage(l.logger, l.peerID, msgID, err)
	}
}

// handleUnicast finds the target session for a unicast notification and
// invokes the handler.
func (l *listener) handleUnicast(
	msgID ident.MessageID,
	msg *amqp.Delivery,
	source rinq.Revision,
) error {
	sessID, err := ident.ParseSessionID(msg.RoutingKey)
	if err != nil {
		return err
	}

	sess, _, ok := l.sessions.Get(sessID)
	if !ok {
		return nil
	} else if err != nil {
		return err
	}

	l.handle(
		amqputil.UnpackTrace(l.parentCtx, msg),
		msgID,
		sess,
		rinq.Notification{
			Source:  source,
			Type:    msg.Type,
			Payload: rinq.NewPayloadFromBytes(msg.Body),
		},
	)

	return nil
}

// handleUnicast finds the target sessions for a multicast notification and
// invokes the handlers.
func (l *listener) handleMulticast(
	msgID ident.MessageID,
	msg *amqp.Delivery,
	source rinq.Revision,
) error {
	constraint := rinq.Constraint{}
	for key, value := range msg.Headers {
		if v, ok := value.(string); ok {
			constraint[key] = v
		} else {
			return fmt.Errorf("constraint key %s contains non-string value", key)
		}
	}

	var sessions []rinq.Session

	l.sessions.Each(func(session rinq.Session, catalog localsession.Catalog) {
		_, attrs := catalog.Attrs()
		if attrs.MatchConstraint(constraint) {
			sessions = append(sessions, session)
		}
	})

	if len(sessions) == 0 {
		return nil
	}

	ctx := amqputil.UnpackTrace(l.parentCtx, msg)
	payload := rinq.NewPayloadFromBytes(msg.Body)
	defer payload.Close()

	for _, sess := range sessions {
		l.handle(
			ctx,
			msgID,
			sess,
			rinq.Notification{
				Source:      source,
				Type:        msg.Type,
				Payload:     payload.Clone(),
				IsMulticast: true,
				Constraint:  constraint,
			},
		)
	}

	return nil
}

// handle invokes the notification handler for a specific session, if one is
// present.
func (l *listener) handle(
	ctx context.Context,
	msgID ident.MessageID,
	sess rinq.Session,
	n rinq.Notification,
) {
	l.mutex.RLock()
	handler := l.handlers[sess.ID()]
	l.mutex.RUnlock()

	if handler == nil {
		n.Payload.Close()
	} else {
		handler(ctx, sess, n)
	}
}

func (l *listener) closeChannel() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.channel.Close()
}
