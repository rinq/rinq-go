package notifyamqp

import (
	"context"
	"fmt"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/internal/localsession"
	"github.com/rinq/rinq-go/src/internal/notify"
	"github.com/rinq/rinq-go/src/internal/revisions"
	"github.com/rinq/rinq-go/src/internal/service"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/amqputil"
	"github.com/streadway/amqp"
)

type listener struct {
	service.Service
	sm *service.StateMachine

	peerID    ident.PeerID
	preFetch  uint
	sessions  *localsession.Store
	revisions revisions.Store
	logger    rinq.Logger
	tracer    opentracing.Tracer

	parentCtx context.Context // parent of all contexts passed to handlers
	cancelCtx func()          // cancels parentCtx when the server stops

	// state-machine data
	channel    *amqp.Channel        // channel used for consuming
	namespaces map[string]uint      // map of namespace to listener count
	deliveries <-chan amqp.Delivery // incoming notifications
	amqpClosed chan *amqp.Error
	pending    uint // number of notifications currently being handled

	mutex    sync.RWMutex // guards handlers so handler can be read in dispatch() goroutine
	handlers map[ident.SessionID]map[string]rinq.NotificationHandler
}

// newListener creates, starts and returns a new listener.
func newListener(
	peerID ident.PeerID,
	preFetch uint,
	sessions *localsession.Store,
	revs revisions.Store,
	channel *amqp.Channel,
	logger rinq.Logger,
	tracer opentracing.Tracer,
) (notify.Listener, error) {
	l := &listener{
		peerID:    peerID,
		preFetch:  preFetch,
		sessions:  sessions,
		revisions: revs,
		logger:    logger,
		tracer:    tracer,

		channel:    channel,
		namespaces: map[string]uint{},
		amqpClosed: make(chan *amqp.Error, 1),

		handlers: map[ident.SessionID]map[string]rinq.NotificationHandler{},
	}

	l.sm = service.NewStateMachine(l.run, l.finalize)
	l.Service = l.sm

	if err := l.initialize(); err != nil {
		return nil, err
	}

	go l.sm.Run()

	return l, nil
}

func (l *listener) Listen(id ident.SessionID, ns string, h rinq.NotificationHandler) (added bool, err error) {
	err = l.sm.Do(func() error {
		l.mutex.Lock()
		defer l.mutex.Unlock()

		handlers, ok := l.handlers[id]
		if !ok {
			handlers = map[string]rinq.NotificationHandler{}
			l.handlers[id] = handlers
		}

		_, ok = handlers[ns]
		handlers[ns] = h

		if ok {
			return nil
		}

		added = true

		return l.bind(ns)
	})

	return
}

func (l *listener) Unlisten(id ident.SessionID, ns string) (removed bool, err error) {
	err = l.sm.Do(func() error {
		l.mutex.Lock()
		defer l.mutex.Unlock()

		handlers, ok := l.handlers[id]
		if !ok {
			return nil
		}

		_, ok = handlers[ns]
		if !ok {
			return nil
		}

		delete(handlers, ns)
		removed = true

		return l.unbind(ns)
	})

	return
}

func (l *listener) UnlistenAll(id ident.SessionID) error {
	return l.sm.Do(func() error {
		l.mutex.Lock()
		defer l.mutex.Unlock()

		handlers := l.handlers[id]
		delete(l.handlers, id)

		for ns := range handlers {
			if err := l.unbind(ns); err != nil {
				return err
			}
		}

		return nil
	})
}

func (l *listener) bind(ns string) error {
	count := l.namespaces[ns]
	l.namespaces[ns] = count + 1

	if count != 0 {
		return nil
	}

	queue := notifyQueue(l.peerID)

	if err := l.channel.QueueBind(
		queue,
		unicastRoutingKey(ns, l.peerID),
		unicastExchange,
		false, // noWait
		nil,   // args
	); err != nil {
		return err
	}

	return l.channel.QueueBind(
		queue,
		ns,
		multicastExchange,
		false, // noWait
		nil,   // args
	)
}

func (l *listener) unbind(ns string) error {
	count := l.namespaces[ns] - 1
	l.namespaces[ns] = count

	if count != 0 {
		return nil
	}

	queue := notifyQueue(l.peerID)

	if err := l.channel.QueueUnbind(
		queue,
		unicastRoutingKey(ns, l.peerID),
		unicastExchange,
		nil, // args
	); err != nil {
		return err
	}

	return l.channel.QueueUnbind(
		queue,
		ns,
		multicastExchange,
		nil, // args
	)
}

// initialize prepares the AMQP channel
func (l *listener) initialize() error {
	l.channel.NotifyClose(l.amqpClosed)

	queue := notifyQueue(l.peerID)

	// if _, err := l.channel.QueueDeclare(
	// 	queue,
	// 	false, // durable
	// 	false, // autoDelete
	// 	true,  // exclusive,
	// 	false, // noWait
	// 	nil,   // args
	// ); err != nil {
	// 	panic(err)
	// }

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

	if err != nil {
		panic(err)
	}

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

		case req := <-l.sm.Commands:
			l.sm.Execute(req)

		case <-l.sm.Graceful:
			return l.stopConsuming, nil

		case <-l.sm.Forceful:
			return nil, nil

		case err := <-l.amqpClosed:
			return nil, err
		}
	}
}

// stopConsuming is the first state entered when a graceful stop is requested.
func (l *listener) stopConsuming() (service.State, error) {
	logListenerStopping(l.logger, l.peerID, l.pending)

	queue := notifyQueue(l.peerID)
	if err := l.channel.Cancel(queue, false); err != nil { // false = wait for response
		return nil, err
	}

	// reject any messages that have already been delivered
	// l.deliveries is closed by the call to Cancel() above
	for msg := range l.deliveries {
		if err := msg.Reject(false); err != nil { // false = don't requeue
			return nil, err
		}
	}

	return l.waitForHandlers, nil
}

// waitForHandlers is the second phase of a graceful stop. It waits for any
// pending notification handlers to complete.
func (l *listener) waitForHandlers() (service.State, error) {
	for l.pending > 0 {
		select {
		case req := <-l.sm.Commands:
			l.sm.Execute(req)

		case <-l.sm.Forceful:
			return nil, nil
		}
	}

	return nil, nil
}

// finalize is the state-machine finalizer, it is called immediately before the
// Done() channel is closed.
func (l *listener) finalize(err error) error {
	l.cancelCtx()
	logListenerStop(l.logger, l.peerID, err)

	closeErr := l.channel.Close()

	// only report the closeErr if there's no causal error.
	if err == nil {
		return closeErr
	}

	return err
}

// dispatch validates an incoming notification and dispatches it the
// appropriate handler.
func (l *listener) dispatch(msg *amqp.Delivery) {
	defer l.sm.DoGraceful(func() error {
		l.pending--
		return nil
	})

	// create a prototype notification that is cloned for each handler
	proto := &rinq.Notification{}

	var err error
	proto.ID, err = ident.ParseMessageID(msg.MessageId)
	if err != nil {
		_ = msg.Reject(false) // false = don't requeue
		logInvalidMessageID(l.logger, l.peerID, msg.MessageId)
	}

	defer func() {
		if err == nil {
			_ = msg.Ack(false) // false = single message
		} else {
			_ = msg.Reject(false) // false = don't requeue
			logIgnoredMessage(l.logger, l.peerID, proto.ID, err)
		}
	}()

	// find the source session revision
	proto.Source, err = l.revisions.GetRevision(proto.ID.Ref)
	if err != nil {
		return
	}

	proto.Namespace, proto.Type, proto.Payload, err = unpackCommonAttributes(msg)
	if err != nil {
		return
	}
	defer proto.Payload.Close()

	var sessions []rinq.Session

	switch msg.Exchange {
	case unicastExchange:
		sessions, err = l.findUnicastTarget(proto, msg)
	case multicastExchange:
		proto.IsMulticast = true
		sessions, err = l.findMulticastTargets(proto, msg)
	default:
		err = fmt.Errorf("delivery via '%s' exchange is not expected", msg.Exchange)
	}
	if err != nil {
		return
	}

	ctx := amqputil.UnpackTrace(l.parentCtx, msg)

	spanOpts, err := unpackSpanOptions(msg, l.tracer)
	if err != nil {
		return
	}

	for _, sess := range sessions {
		l.handle(
			ctx,
			sess,
			proto,
			spanOpts,
		)
	}
}

// findUnicastTarget returns the session that should receive the unicast
// notification n.
func (l *listener) findUnicastTarget(
	n *rinq.Notification,
	msg *amqp.Delivery,
) ([]rinq.Session, error) {
	var sessID ident.SessionID
	sessID, err := unpackTarget(msg)
	if err != nil {
		return nil, err
	}

	if sess, ok := l.sessions.Get(sessID); ok {
		return []rinq.Session{sess}, nil
	}

	return nil, nil
}

// findMulticastTargets returns the sessions that should receive the multicast
// notification n.
func (l *listener) findMulticastTargets(
	n *rinq.Notification,
	msg *amqp.Delivery,
) (
	sessions []rinq.Session,
	err error,
) {
	n.Constraint, err = unpackConstraint(msg)
	if err != nil {
		return
	}

	l.sessions.Each(
		func(session *localsession.Session) {
			_, attrs := session.Attrs()
			if attrs.MatchConstraint(n.Namespace, n.Constraint) {
				sessions = append(sessions, session)
			}
		},
	)

	return
}

// handle invokes the notification handler for a specific session, if one is
// present.
func (l *listener) handle(
	ctx context.Context,
	sess rinq.Session,
	proto *rinq.Notification,
	spanOpts []opentracing.StartSpanOption,
) {
	l.mutex.RLock()
	h := l.handlers[sess.ID()][proto.Namespace]
	l.mutex.RUnlock()

	if h != nil {
		n := *proto
		n.Payload = n.Payload.Clone()

		span := l.tracer.StartSpan("", spanOpts...)
		defer span.Finish()

		h(
			opentracing.ContextWithSpan(ctx, span),
			sess,
			n,
		)
	}
}
