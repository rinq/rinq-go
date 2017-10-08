package notifyamqp

import (
	"context"
	"fmt"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
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
	preFetch  uint
	sessions  localsession.Store
	revisions revision.Store
	logger    rinq.Logger
	tracer    opentracing.Tracer

	parentCtx context.Context // parent of all contexts passed to handlers
	cancelCtx func()          // cancels parentCtx when the server stops

	mutex      sync.RWMutex
	channel    *amqp.Channel   // channel used for consuming
	namespaces map[string]uint // map of namespace to listener count
	handlers   map[ident.SessionID]map[string]notify.Handler

	deliveries <-chan amqp.Delivery // incoming notifications
	handled    chan struct{}        // signals a notification has been handled
	amqpClosed chan *amqp.Error

	// state-machine data
	pending uint // number of notifications currently being handled
}

// newListener creates, starts and returns a new listener.
func newListener(
	peerID ident.PeerID,
	preFetch uint,
	sessions localsession.Store,
	revisions revision.Store,
	channel *amqp.Channel,
	logger rinq.Logger,
	tracer opentracing.Tracer,
) (notify.Listener, error) {
	l := &listener{
		peerID:    peerID,
		preFetch:  preFetch,
		sessions:  sessions,
		revisions: revisions,
		logger:    logger,
		tracer:    tracer,
		channel:   channel,

		namespaces: map[string]uint{},
		handlers:   map[ident.SessionID]map[string]notify.Handler{},

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

func (l *listener) Listen(id ident.SessionID, ns string, handler notify.Handler) (bool, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	handlers, ok := l.handlers[id]
	if !ok {
		handlers = map[string]notify.Handler{}
		l.handlers[id] = handlers
	}

	_, ok = handlers[ns]
	handlers[ns] = handler

	if ok {
		return false, nil
	}

	return true, l.bindQueues(ns)
}

func (l *listener) Unlisten(id ident.SessionID, ns string) (bool, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	handlers, ok := l.handlers[id]
	if !ok {
		return false, nil
	}

	_, ok = handlers[ns]
	if !ok {
		return false, nil
	}

	delete(handlers, ns)

	if len(handlers) == 0 {
		delete(l.handlers, id)
	}

	return true, l.unbindQueues(ns)
}

func (l *listener) UnlistenAll(id ident.SessionID) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	handlers := l.handlers[id]
	delete(l.handlers, id)

	for ns := range handlers {
		if err := l.unbindQueues(ns); err != nil {
			return err
		}
	}

	return nil
}

func (l *listener) bindQueues(ns string) error {
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

func (l *listener) unbindQueues(ns string) error {
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
	var err error

	defer func() {
		if err == nil {
			_ = msg.Ack(false) // false = single message
		} else {
			_ = msg.Reject(false) // false = don't requeue
			logInvalidMessageID(l.logger, l.peerID, msg.MessageId)
		}

		select {
		case l.handled <- struct{}{}:
		case <-l.sm.Forceful:
		}
	}()

	msgID, err := ident.ParseMessageID(msg.MessageId)
	if err != nil {
		return
	}

	// create a prototype notification that is cloned for each handler
	proto := &rinq.Notification{}

	// find the source session revision
	proto.Source, err = l.revisions.GetRevision(msgID.Ref)
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
			msgID,
			sess,
			proto,
			spanOpts,
		)
	}
}

// findUnicastTarget returns the session that should receive the unicast
// notificaton n.
func (l *listener) findUnicastTarget(
	n *rinq.Notification,
	msg *amqp.Delivery,
) ([]rinq.Session, error) {
	var sessID ident.SessionID
	sessID, err := unpackTarget(msg)
	if err != nil {
		return nil, err
	}

	if sess, _, ok := l.sessions.Get(sessID); ok {
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
		func(session rinq.Session, catalog localsession.Catalog) {
			_, attrs := catalog.Attrs()
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
	msgID ident.MessageID,
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
			msgID,
			sess,
			n,
		)
	}
}

func (l *listener) closeChannel() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.channel.Close()
}
