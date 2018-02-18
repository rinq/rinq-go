package localsession

import (
	"context"

	"github.com/jmalloc/twelf/src/twelf"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/internal/logging"
	"github.com/rinq/rinq-go/src/internal/opentr"
	"github.com/rinq/rinq-go/src/internal/revisions"
	"github.com/rinq/rinq-go/src/internal/transport"
	"github.com/rinq/rinq-go/src/rinq"
)

// receiver accepts incoming notifications and invokes the associated handler.
// Handlers are invoked in order on a single goroutine.
type receiver struct {
	revs   revisions.Store
	tracer opentracing.Tracer
	logger twelf.Logger

	session  *Session
	handlers map[string]rinq.NotificationHandler

	ntf  chan transport.InboundNotification
	bind chan binding
	done chan struct{}
}

type binding struct {
	Namespace string
	Handler   rinq.NotificationHandler
}

// newReceiver creates a new receiver for the given session. cap is the size
// of the notification input buffer.
func newReceiver(
	revs revisions.Store,
	tracer opentracing.Tracer,
	logger twelf.Logger,
	cap int,
) *receiver {
	return &receiver{
		revs:   revs,
		tracer: tracer,
		logger: logger,
		ntf:    make(chan transport.InboundNotification, cap),
		bind:   make(chan binding),
		done:   make(chan struct{}),
	}
}

// Run accepts messages and invokes the associated handlers until ctx is
// canceled.
func (r *receiver) Run(s *Session) {
	defer close(r.done)

	r.session = s
	r.handlers = map[string]rinq.NotificationHandler{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case <-s.Done():
			return
		case b := <-r.bind:
			r.setHandler(b.Namespace, b.Handler)
		case n := <-r.ntf:
			r.handle(ctx, n)
		}
	}
}

// Accept informs the receiver of a notification.
func (r *receiver) Accept(n transport.InboundNotification) {
	select {
	case r.ntf <- n:
	case <-r.done:
		n.Done()
	}
}

// SetHandler associates h with the ns namespace. h may be nil to indicate that
// there is no handler for ns.
func (r *receiver) SetHandler(ns string, h rinq.NotificationHandler) {
	select {
	case r.bind <- binding{ns, h}:
	case <-r.done:
	}
}

func (r *receiver) setHandler(ns string, h rinq.NotificationHandler) {
	_, ok := r.handlers[ns]

	ref, _ := r.session.Attrs()

	if h == nil {
		delete(r.handlers, ns)
		if ok {
			logging.NotificationUnsubscribe(r.logger, ref, ns)
		}
	} else {
		r.handlers[ns] = h
		if !ok {
			logging.NotificationSubscribe(r.logger, ref, ns)
		}
	}
}

func (r *receiver) handle(ctx context.Context, n transport.InboundNotification) {
	defer n.Done()

	h, ok := r.handlers[n.Notification.Namespace]
	if !ok {
		return
	}

	ref, attrs := r.session.Attrs()

	if n.IsMulticast && !attrs.MatchConstraint(n.Namespace, n.MulticastConstraint) {
		return
	}

	source, err := r.revs.GetRevision(n.ID.Ref)
	if err != nil {
		logging.IgnoredNotification(r.logger, ref, n.Notification, err)
		return
	}

	span := opentr.ReceivedNotification(r.tracer, ref, n.Notification)
	defer span.Finish()

	logging.ReceivedNotification(r.logger, ref, n.Notification)

	h(ctx, r.session, rinq.Notification{
		ID:          n.ID,
		Source:      source,
		Namespace:   n.Namespace,
		Type:        n.Type,
		Payload:     n.Payload.Clone(),
		IsMulticast: n.IsMulticast,
		Constraint:  n.MulticastConstraint,
	})
}
