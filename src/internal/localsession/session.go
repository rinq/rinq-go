package localsession

import (
	"context"
	"errors"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/rinq/rinq-go/src/internal/attributes"
	"github.com/rinq/rinq-go/src/internal/command"
	"github.com/rinq/rinq-go/src/internal/namespaces"
	"github.com/rinq/rinq-go/src/internal/notify"
	"github.com/rinq/rinq-go/src/internal/opentr"
	"github.com/rinq/rinq-go/src/internal/revisions"
	"github.com/rinq/rinq-go/src/internal/x/syncx"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/constraint"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/trace"
)

// Session is an extension of the rinq.Session interface that includes
// operations used by Rinq internals.
type Session interface {
	rinq.Session

	// At returns a revision representing the state at a specific revision
	// number. The revision can not be newer than the current session-ref.
	At(rev ident.Revision) (rinq.Revision, error)

	// Attrs returns all attributes at the most recent revision.
	Attrs() (ident.Ref, attributes.Catalog)

	// AttrsIn returns all attributes in the ns namespace at the most recent revision.
	AttrsIn(ns string) (ident.Ref, attributes.VTable)

	// TryUpdate adds or updates attributes in the ns namespace of the attribute
	// table and returns the new head revision.
	//
	// The operation fails if ref is not the current session-ref, attrs includes
	// changes to frozen attributes, or the session has been destroyed.
	TryUpdate(rev ident.Revision, ns string, attrs attributes.List) (rinq.Revision, *attributes.Diff, error)

	// TryClear updates all attributes in the ns namespace of the attribute
	// table to the empty string and returns the new head revision.
	//
	// The operation fails if ref is not the current session-ref, there are any
	// frozen attributes, or the session has been destroyed.
	TryClear(rev ident.Revision, ns string) (rinq.Revision, *attributes.Diff, error)

	// TryDestroy destroys the session, preventing further updates.
	//
	// The operation fails if ref is not the current session-ref. It is not an
	// error to destroy an already-destroyed session.
	//
	// first is true if this call caused the session to be destroyed.
	TryDestroy(rev ident.Revision) (first bool, err error)
}

type session struct {
	invoker  command.Invoker
	notifier notify.Notifier
	listener notify.Listener
	logger   rinq.Logger
	tracer   opentracing.Tracer

	mutex       sync.RWMutex
	ref         ident.Ref
	msgSeq      uint32
	isDestroyed bool
	attrs       attributes.Catalog
	calls       sync.WaitGroup
	done        chan struct{}
}

// NewSession returns a new local session.
func NewSession(
	id ident.SessionID,
	invoker command.Invoker,
	notifier notify.Notifier,
	listener notify.Listener,
	logger rinq.Logger,
	tracer opentracing.Tracer,
) Session {
	logCreated(logger, id)

	return &session{
		invoker:  invoker,
		notifier: notifier,
		listener: listener,
		logger:   logger,
		tracer:   tracer,

		ref:  id.At(0),
		done: make(chan struct{}),
	}
}

func (s *session) ID() ident.SessionID {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.ref.ID
}

func (s *session) CurrentRevision() rinq.Revision {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.isDestroyed {
		return revisions.Closed(s.ref.ID)
	}

	return &revision{s, s.ref, s.attrs, s.logger}
}

func (s *session) Call(ctx context.Context, ns, cmd string, out *rinq.Payload) (*rinq.Payload, error) {
	namespaces.MustValidate(ns)

	unlock := syncx.Lock(&s.mutex)
	defer unlock()

	if s.isDestroyed {
		return nil, rinq.NotFoundError{ID: s.ref.ID}
	}

	msgID := s.nextMessageID()
	attrs := s.attrs // capture for logging/tracing while mutex is locked

	s.calls.Add(1)
	defer s.calls.Done()

	// do not hold the lock for the duration of the call, as this would prevent
	// the handler of the call querying or modifying this session.
	unlock()

	span, ctx := opentr.ChildOf(ctx, s.tracer, ext.SpanKindRPCClient)
	defer span.Finish()

	opentr.SetupCommand(span, msgID, ns, cmd)
	opentr.LogInvokerCall(span, attrs, out)

	start := time.Now()
	traceID, in, err := s.invoker.CallBalanced(ctx, msgID, ns, cmd, out)
	elapsed := time.Since(start) / time.Millisecond

	if err == nil {
		opentr.LogInvokerSuccess(span, in)
	} else {
		opentr.LogInvokerError(span, err)
	}

	logCall(s.logger, msgID, ns, cmd, elapsed, out, in, err, traceID)

	return in, err
}

func (s *session) CallAsync(ctx context.Context, ns, cmd string, out *rinq.Payload) (ident.MessageID, error) {
	namespaces.MustValidate(ns)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isDestroyed {
		return ident.MessageID{}, rinq.NotFoundError{ID: s.ref.ID}
	}

	msgID := s.nextMessageID()

	span, ctx := opentr.ChildOf(ctx, s.tracer, ext.SpanKindRPCClient)
	defer span.Finish()

	opentr.SetupCommand(span, msgID, ns, cmd)
	opentr.LogInvokerCallAsync(span, s.attrs, out)

	traceID, err := s.invoker.CallBalancedAsync(ctx, msgID, ns, cmd, out)

	if err != nil {
		opentr.LogInvokerError(span, err)
	}

	logAsyncRequest(s.logger, msgID, ns, cmd, out, err, traceID)

	return msgID, err
}

// SetAsyncHandler sets the asynchronous call handler.
//
// h is invoked for each command response received to a command request made
// with CallAsync().
func (s *session) SetAsyncHandler(h rinq.AsyncHandler) error {
	// it is important that this lock is acquired for the duration of the call
	// to s.invoker.SetAsyncHandler(), to ensure that it is serialized with
	// the similar call in s.destroy() which sets the handler to nil.
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.isDestroyed {
		return rinq.NotFoundError{ID: s.ref.ID}
	}

	s.invoker.SetAsyncHandler(
		s.ref.ID,
		func(
			ctx context.Context,
			sess rinq.Session,
			msgID ident.MessageID,
			ns string,
			cmd string,
			in *rinq.Payload,
			err error,
		) {
			span := opentracing.SpanFromContext(ctx)
			opentr.SetupCommand(span, msgID, ns, cmd)

			if err == nil {
				opentr.LogInvokerSuccess(span, in)
			} else {
				opentr.LogInvokerError(span, err)
			}

			logAsyncResponse(ctx, s.logger, msgID, ns, cmd, in, err)

			h(ctx, sess, msgID, ns, cmd, in, err)
		},
	)

	return nil
}

func (s *session) Execute(ctx context.Context, ns, cmd string, p *rinq.Payload) error {
	namespaces.MustValidate(ns)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isDestroyed {
		return rinq.NotFoundError{ID: s.ref.ID}
	}

	msgID := s.nextMessageID()

	span, ctx := opentr.ChildOf(ctx, s.tracer, ext.SpanKindRPCClient)
	defer span.Finish()

	opentr.SetupCommand(span, msgID, ns, cmd)
	opentr.LogInvokerCallAsync(span, s.attrs, p)

	traceID, err := s.invoker.ExecuteBalanced(ctx, msgID, ns, cmd, p)

	if err != nil {
		opentr.LogInvokerError(span, err)
	}

	// TODO: move to function
	if err == nil {
		s.logger.Log(
			"%s executed '%s::%s' command (%d/o) [%s]",
			msgID.ShortString(),
			ns,
			cmd,
			p.Len(),
			traceID,
		)
	}

	return err
}

func (s *session) Notify(ctx context.Context, ns, t string, target ident.SessionID, p *rinq.Payload) error {
	namespaces.MustValidate(ns)
	ident.MustValidate(target)
	if target.Seq == 0 {
		panic("can not send notifications to the zero-session")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isDestroyed {
		return rinq.NotFoundError{ID: s.ref.ID}
	}

	msgID := s.nextMessageID()

	span, ctx := opentr.ChildOf(ctx, s.tracer, ext.SpanKindProducer)
	defer span.Finish()

	opentr.SetupNotification(span, msgID, ns, t)
	opentr.LogNotifierUnicast(span, s.attrs, target, p)

	traceID, err := s.notifier.NotifyUnicast(ctx, msgID, target, ns, t, p)

	if err != nil {
		opentr.LogNotifierError(span, err)
	}

	// TODO: move to function
	if err == nil {
		s.logger.Log(
			"%s sent '%s::%s' notification to %s (%d/o) [%s]",
			msgID.ShortString(),
			ns,
			t,
			target.ShortString(),
			p.Len(),
			traceID,
		)
	}

	return err
}

func (s *session) NotifyMany(ctx context.Context, ns, t string, con constraint.Constraint, p *rinq.Payload) error {
	namespaces.MustValidate(ns)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isDestroyed {
		return rinq.NotFoundError{ID: s.ref.ID}
	}

	msgID := s.nextMessageID()

	span, ctx := opentr.ChildOf(ctx, s.tracer, ext.SpanKindProducer)
	defer span.Finish()

	opentr.SetupNotification(span, msgID, ns, t)
	opentr.LogNotifierMulticast(span, s.attrs, con, p)

	traceID, err := s.notifier.NotifyMulticast(ctx, msgID, con, ns, t, p)

	if err != nil {
		opentr.LogNotifierError(span, err)
	}

	// TODO: move to function
	if err == nil {
		s.logger.Log(
			"%s sent '%s::%s' notification to sessions matching %s (%d/o) [%s]",
			msgID.ShortString(),
			ns,
			t,
			con,
			p.Len(),
			traceID,
		)
	}

	return err
}

func (s *session) Listen(ns string, h rinq.NotificationHandler) error {
	namespaces.MustValidate(ns)
	if h == nil {
		panic("handler must not be nil")
	}

	// it is important that this lock is acquired for the duration of the call
	// to s.listener.Listen(), to ensure that it is serialized with the call
	// to s.listener.UnlistenAll() in s.destroy().
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.isDestroyed {
		return rinq.NotFoundError{ID: s.ref.ID}
	}

	changed, err := s.listener.Listen(
		s.ref.ID,
		ns,
		func(
			ctx context.Context,
			target rinq.Session,
			n rinq.Notification,
		) {
			s.mutex.RLock()
			ref := s.ref
			s.mutex.RUnlock()

			span := opentracing.SpanFromContext(ctx)
			opentr.SetupNotification(span, n.ID, n.Namespace, n.Type)
			opentr.LogListenerReceived(span, ref, n)

			// TODO: move to function
			s.logger.Log(
				"%s received '%s::%s' notification from %s (%d/i) [%s]",
				ref.ShortString(),
				n.Namespace,
				n.Type,
				n.ID.Ref.ShortString(),
				n.Payload.Len(),
				trace.Get(ctx),
			)

			h(ctx, target, n)
		},
	)

	if err != nil {
		return err
	} else if changed && s.logger.IsDebug() {
		s.logger.Log(
			"%s started listening for notifications in '%s' namespace",
			s.ref.ShortString(),
			ns,
		)
	}

	return nil
}

func (s *session) Unlisten(ns string) error {
	namespaces.MustValidate(ns)

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.isDestroyed {
		return rinq.NotFoundError{ID: s.ref.ID}
	}

	changed, err := s.listener.Unlisten(s.ref.ID, ns)

	if err != nil {
		return err
	} else if changed && s.logger.IsDebug() {
		s.logger.Log(
			"%s stopped listening for notifications in '%s' namespace",
			s.ref.ShortString(),
			ns,
		)
	}

	return nil
}

func (s *session) At(rev ident.Revision) (rinq.Revision, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.ref.Rev < rev {
		return nil, errors.New("revision is from the future")
	}

	return &revision{
		s,
		s.ref,
		s.attrs,
		s.logger,
	}, nil
}

func (s *session) Attrs() (ident.Ref, attributes.Catalog) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.ref, s.attrs
}

func (s *session) AttrsIn(ns string) (ident.Ref, attributes.VTable) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.ref, s.attrs[ns]
}

func (s *session) TryUpdate(rev ident.Revision, ns string, attrs attributes.List) (rinq.Revision, *attributes.Diff, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isDestroyed {
		return nil, nil, rinq.NotFoundError{ID: s.ref.ID}
	}

	if rev != s.ref.Rev {
		return nil, nil, rinq.StaleUpdateError{Ref: s.ref.ID.At(rev)}
	}

	nextRev := rev + 1
	nextAttrs := s.attrs[ns].Clone()
	diff := attributes.NewDiff(ns, nextRev)

	for _, attr := range attrs {
		entry, exists := nextAttrs[attr.Key]

		if attr.Value == entry.Value && attr.IsFrozen == entry.IsFrozen {
			continue
		}

		if entry.IsFrozen {
			return nil, nil, rinq.FrozenAttributesError{Ref: s.ref.ID.At(rev)}
		}

		entry.Attr = attr
		entry.UpdatedAt = nextRev
		if !exists {
			entry.CreatedAt = nextRev
		}

		nextAttrs[attr.Key] = entry
		diff.Append(entry)
	}

	s.ref.Rev = nextRev
	s.msgSeq = 0

	if !diff.IsEmpty() {
		s.attrs = s.attrs.WithNamespace(ns, nextAttrs)
	}

	return &revision{
		s,
		s.ref,
		s.attrs,
		s.logger,
	}, diff, nil
}

func (s *session) TryClear(rev ident.Revision, ns string) (rinq.Revision, *attributes.Diff, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isDestroyed {
		return nil, nil, rinq.NotFoundError{ID: s.ref.ID}
	}

	if rev != s.ref.Rev {
		return nil, nil, rinq.StaleUpdateError{Ref: s.ref.ID.At(rev)}
	}

	attrs := s.attrs[ns]
	nextRev := rev + 1
	nextAttrs := attributes.VTable{}
	diff := attributes.NewDiff(ns, nextRev)

	for _, entry := range attrs {
		if entry.Value != "" {
			if entry.IsFrozen {
				return nil, nil, rinq.FrozenAttributesError{Ref: s.ref.ID.At(rev)}
			}

			entry.Value = ""
			entry.UpdatedAt = nextRev
			diff.Append(entry)
		}

		nextAttrs[entry.Key] = entry
	}

	s.ref.Rev = nextRev
	s.msgSeq = 0

	if !diff.IsEmpty() {
		s.attrs = s.attrs.WithNamespace(ns, nextAttrs)
	}

	return &revision{
		s,
		s.ref,
		s.attrs,
		s.logger,
	}, diff, nil
}

func (s *session) TryDestroy(rev ident.Revision) (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if rev != s.ref.Rev {
		return false, rinq.StaleUpdateError{Ref: s.ref.ID.At(rev)}
	}

	if s.isDestroyed {
		return false, nil
	}

	s.destroy()

	return true, nil
}

func (s *session) Destroy() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.isDestroyed {
		s.destroy()
		logSessionDestroy(s.logger, s.ref, s.attrs, "")
	}
}

func (s *session) destroy() {
	s.isDestroyed = true

	s.invoker.SetAsyncHandler(s.ref.ID, nil)
	_ = s.listener.UnlistenAll(s.ref.ID)

	go func() {
		s.calls.Wait()
		close(s.done)
	}()
}

func (s *session) Done() <-chan struct{} {
	return s.done
}

// nextMessageID returns a new unique message ID generated from the current
// session-ref, and the attributes as they existed at that ref.
func (s *session) nextMessageID() ident.MessageID {
	s.msgSeq++
	return s.ref.Message(s.msgSeq)
}
