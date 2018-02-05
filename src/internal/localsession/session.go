package localsession

import (
	"context"
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

// Session is the implementation of rinq.Session.
//
// The implementation is split into two files. This file contains methods that
// are declared in rinq.Session, whereas the session_state.go file contains a
// lower-level API for manipulating the session state which is used throughout
// the Rinq internals.
type Session struct {
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
) *Session {
	logCreated(logger, id)

	return &Session{
		invoker:  invoker,
		notifier: notifier,
		listener: listener,
		logger:   logger,
		tracer:   tracer,

		ref:  id.At(0),
		done: make(chan struct{}),
	}
}

// ID implements rinq.Session.ID()
func (s *Session) ID() ident.SessionID {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.ref.ID
}

// CurrentRevision implements rinq.Session.CurrentRevision()
func (s *Session) CurrentRevision() rinq.Revision {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.isDestroyed {
		return revisions.Closed(s.ref.ID)
	}

	return &revision{s.ref, s, s.attrs, s.logger}
}

// Call implements rinq.Session.Call()
func (s *Session) Call(ctx context.Context, ns, cmd string, out *rinq.Payload) (*rinq.Payload, error) {
	namespaces.MustValidate(ns)

	unlock := syncx.Lock(&s.mutex)
	defer unlock()

	if s.isDestroyed {
		return nil, rinq.NotFoundError{ID: s.ref.ID}
	}

	msgID, traceID := s.nextMessageID(ctx)
	attrs := s.attrs // capture for logging/tracing while mutex is locked

	s.calls.Add(1)
	defer s.calls.Done()

	// do not hold the lock for the duration of the call, as this would prevent
	// the handler of the call querying or modifying this session.
	unlock()

	span, ctx := opentr.ChildOf(ctx, s.tracer, ext.SpanKindRPCClient)
	defer span.Finish()

	opentr.SetupCommand(span, msgID, ns, cmd)
	opentr.AddTraceID(span, traceID)
	opentr.LogInvokerCall(span, attrs, out)

	start := time.Now()
	in, err := s.invoker.CallBalanced(ctx, msgID, traceID, ns, cmd, out)
	elapsed := time.Since(start) / time.Millisecond

	if err == nil {
		opentr.LogInvokerSuccess(span, in)
	} else {
		opentr.LogInvokerError(span, err)
	}

	logCall(s.logger, msgID, ns, cmd, elapsed, out, in, err, traceID)

	return in, err
}

// CallAsync implements rinq.Session.CallAsync()
func (s *Session) CallAsync(ctx context.Context, ns, cmd string, out *rinq.Payload) (ident.MessageID, error) {
	namespaces.MustValidate(ns)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isDestroyed {
		return ident.MessageID{}, rinq.NotFoundError{ID: s.ref.ID}
	}

	msgID, traceID := s.nextMessageID(ctx)

	span, ctx := opentr.ChildOf(ctx, s.tracer, ext.SpanKindRPCClient)
	defer span.Finish()

	opentr.SetupCommand(span, msgID, ns, cmd)
	opentr.AddTraceID(span, traceID)
	opentr.LogInvokerCallAsync(span, s.attrs, out)

	err := s.invoker.CallBalancedAsync(ctx, msgID, traceID, ns, cmd, out)

	if err != nil {
		opentr.LogInvokerError(span, err)
	}

	logAsyncRequest(s.logger, msgID, ns, cmd, out, err, traceID)

	return msgID, err
}

// SetAsyncHandler implements rinq.Session.SetAsyncHandler()
func (s *Session) SetAsyncHandler(h rinq.AsyncHandler) error {
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
			opentr.AddTraceID(span, trace.Get(ctx))

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

// Execute implements rinq.Session.Execute()
func (s *Session) Execute(ctx context.Context, ns, cmd string, p *rinq.Payload) error {
	namespaces.MustValidate(ns)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isDestroyed {
		return rinq.NotFoundError{ID: s.ref.ID}
	}

	msgID, traceID := s.nextMessageID(ctx)

	span, ctx := opentr.ChildOf(ctx, s.tracer, ext.SpanKindRPCClient)
	defer span.Finish()

	opentr.SetupCommand(span, msgID, ns, cmd)
	opentr.AddTraceID(span, traceID)
	opentr.LogInvokerExecute(span, s.attrs, p)

	err := s.invoker.ExecuteBalanced(ctx, msgID, traceID, ns, cmd, p)

	if err != nil {
		opentr.LogInvokerError(span, err)
	}

	logExecute(s.logger, msgID, ns, cmd, p, err, traceID)

	return err
}

// Notify implements rinq.Session.Notify()
func (s *Session) Notify(ctx context.Context, ns, t string, target ident.SessionID, p *rinq.Payload) error {
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

	msgID, traceID := s.nextMessageID(ctx)

	span, ctx := opentr.ChildOf(ctx, s.tracer, ext.SpanKindProducer)
	defer span.Finish()

	opentr.SetupNotification(span, msgID, ns, t)
	opentr.AddTraceID(span, traceID)
	opentr.LogNotifierUnicast(span, s.attrs, target, p)

	err := s.notifier.NotifyUnicast(ctx, msgID, traceID, target, ns, t, p)

	if err != nil {
		opentr.LogNotifierError(span, err)
	}

	logNotify(s.logger, msgID, ns, t, target, p, err, traceID)

	return err
}

// NotifyMany implements rinq.Session.NotifyMany()
func (s *Session) NotifyMany(ctx context.Context, ns, t string, con constraint.Constraint, p *rinq.Payload) error {
	namespaces.MustValidate(ns)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isDestroyed {
		return rinq.NotFoundError{ID: s.ref.ID}
	}

	msgID, traceID := s.nextMessageID(ctx)

	span, ctx := opentr.ChildOf(ctx, s.tracer, ext.SpanKindProducer)
	defer span.Finish()

	opentr.SetupNotification(span, msgID, ns, t)
	opentr.AddTraceID(span, traceID)
	opentr.LogNotifierMulticast(span, s.attrs, con, p)

	err := s.notifier.NotifyMulticast(ctx, msgID, traceID, con, ns, t, p)

	if err != nil {
		opentr.LogNotifierError(span, err)
	}

	logNotifyMany(s.logger, msgID, ns, t, con, p, err, traceID)

	return err
}

// Listen implements rinq.Session.Listen()
func (s *Session) Listen(ns string, h rinq.NotificationHandler) error {
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

			traceID := trace.Get(ctx)

			opentr.SetupNotification(span, n.ID, n.Namespace, n.Type)
			opentr.AddTraceID(span, traceID)
			opentr.LogListenerReceived(span, ref, n)

			logNotifyRecv(s.logger, ref, n, traceID)

			h(ctx, target, n)
		},
	)

	if err != nil {
		return err
	} else if changed && s.logger.IsDebug() {
		logListen(s.logger, s.ref, ns)
	}

	return nil
}

// Unlisten implements rinq.Session.Unlisten()
func (s *Session) Unlisten(ns string) error {
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
		logUnlisten(s.logger, s.ref, ns)
	}

	return nil
}

// Destroy implements rinq.Session.Destroy()
func (s *Session) Destroy() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.isDestroyed {
		s.destroy()
		logSessionDestroy(s.logger, s.ref, s.attrs, "")
	}
}

// Done implements rinq.Session.Done()
func (s *Session) Done() <-chan struct{} {
	return s.done
}
