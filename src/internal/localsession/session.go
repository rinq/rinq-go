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
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/constraint"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/trace"
)

// Session is an extension of the rinq.Session interface that includes
// operations used by Rinq internals.
type Session interface {
	rinq.Session

	// Head returns the most recent revision, even if the session has been
	// destroyed. It is conceptually equivalent to s.At(s.Ref().Rev).
	//
	// TODO: can we either adopt this behaviour for Session.CurrentRevision(),
	// or update all callers to use CurrentRevision() as is.
	CurrentRevisionUnsafe() rinq.Revision

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
	TryUpdate(ref ident.Ref, ns string, attrs attributes.List) (rinq.Revision, *attributes.Diff, error)

	// TryClear updates all attributes in the ns namespace of the attribute
	// table to the empty string and returns the new head revision.
	//
	// The operation fails if ref is not the current session-ref, there are any
	// frozen attributes, or the session has been destroyed.
	TryClear(ref ident.Ref, ns string) (rinq.Revision, *attributes.Diff, error)

	// TryDestroy destroys the session, preventing further updates.
	//
	// The operation fails if ref is not the current session-ref. It is not an
	// error to destroy an already-destroyed session.
	TryDestroy(ref ident.Ref) error
}

type session struct {
	id       ident.SessionID
	state    state
	invoker  command.Invoker
	notifier notify.Notifier
	listener notify.Listener
	logger   rinq.Logger
	tracer   opentracing.Tracer
	done     chan struct{}

	// mutex guards Call(), Listen(), Unlisten() and Close() so that Close()
	// waits for pending calls to complete or timeout, and to ensure that it's
	// call to listener.Unlisten() is not "undone" by the user.
	mutex sync.RWMutex
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
	sess := &session{
		id: id,
		state: state{
			ref:       id.At(0),
			destroyed: make(chan struct{}),
			logger:    logger,
		},
		invoker:  invoker,
		notifier: notifier,
		logger:   logger,
		tracer:   tracer,
		listener: listener,
		done:     make(chan struct{}),
	}

	logCreated(logger, sess.state.Ref())

	go func() {
		<-sess.state.Destroyed()
		sess.tearDown()
	}()

	return sess
}

func (s *session) ID() ident.SessionID {
	return s.id
}

func (s *session) CurrentRevision() (rinq.Revision, error) {
	select {
	case <-s.done:
		return nil, rinq.NotFoundError{ID: s.id}
	default:
		return s.state.Head(), nil
	}
}

func (s *session) Call(ctx context.Context, ns, cmd string, out *rinq.Payload) (*rinq.Payload, error) {
	namespaces.MustValidate(ns)

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	select {
	case <-s.done:
		return nil, rinq.NotFoundError{ID: s.id}
	default:
	}

	msgID, attrs := s.state.NextMessageID()

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

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var msgID ident.MessageID

	select {
	case <-s.done:
		return msgID, rinq.NotFoundError{ID: s.id}
	default:
	}

	msgID, attrs := s.state.NextMessageID()

	span, ctx := opentr.ChildOf(ctx, s.tracer, ext.SpanKindRPCClient)
	defer span.Finish()

	opentr.SetupCommand(span, msgID, ns, cmd)
	opentr.LogInvokerCallAsync(span, attrs, out)

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
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	select {
	case <-s.done:
		return rinq.NotFoundError{ID: s.id}
	default:
	}

	s.invoker.SetAsyncHandler(
		s.id,
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

	select {
	case <-s.done:
		return rinq.NotFoundError{ID: s.id}
	default:
	}

	msgID, attrs := s.state.NextMessageID()

	span, ctx := opentr.ChildOf(ctx, s.tracer, ext.SpanKindRPCClient)
	defer span.Finish()

	opentr.SetupCommand(span, msgID, ns, cmd)
	opentr.LogInvokerCallAsync(span, attrs, p)

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

	select {
	case <-s.done:
		return rinq.NotFoundError{ID: s.id}
	default:
	}

	msgID, attrs := s.state.NextMessageID()

	span, ctx := opentr.ChildOf(ctx, s.tracer, ext.SpanKindProducer)
	defer span.Finish()

	opentr.SetupNotification(span, msgID, ns, t)
	opentr.LogNotifierUnicast(span, attrs, target, p)

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

	select {
	case <-s.done:
		return rinq.NotFoundError{ID: s.id}
	default:
	}

	msgID, attrs := s.state.NextMessageID()

	span, ctx := opentr.ChildOf(ctx, s.tracer, ext.SpanKindProducer)
	defer span.Finish()

	opentr.SetupNotification(span, msgID, ns, t)
	opentr.LogNotifierMulticast(span, attrs, con, p)

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

func (s *session) Listen(ns string, handler rinq.NotificationHandler) error {
	namespaces.MustValidate(ns)

	if handler == nil {
		panic("handler must not be nil")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	select {
	case <-s.done:
		return rinq.NotFoundError{ID: s.id}
	default:
	}

	changed, err := s.listener.Listen(
		s.id,
		ns,
		func(
			ctx context.Context,
			target rinq.Session,
			n rinq.Notification,
		) {
			ref := s.state.Ref()

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

			handler(ctx, target, n)
		},
	)

	if err != nil {
		return err
	} else if changed && s.logger.IsDebug() {
		s.logger.Log(
			"%s started listening for notifications in '%s' namespace",
			s.state.Ref().ShortString(),
			ns,
		)
	}

	return nil
}

func (s *session) Unlisten(ns string) error {
	namespaces.MustValidate(ns)

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	select {
	case <-s.done:
		return rinq.NotFoundError{ID: s.id}
	default:
	}

	changed, err := s.listener.Unlisten(s.id, ns)

	if err != nil {
		return err
	} else if changed && s.logger.IsDebug() {
		s.logger.Log(
			"%s stopped listening for notifications in '%s' namespace",
			s.state.Ref().ShortString(),
			ns,
		)
	}

	return nil
}

func (s *session) Destroy() {
	if s.state.ForceDestroy() {
		logSessionDestroy(s.logger, &s.state, "")
	}
}

func (s *session) tearDown() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	select {
	case <-s.done:
	default:
		close(s.done)
		s.invoker.SetAsyncHandler(s.id, nil)
		_ = s.listener.UnlistenAll(s.id)
	}
}

func (s *session) Done() <-chan struct{} {
	return s.done
}

func (s *session) CurrentRevisionUnsafe() rinq.Revision {
	return s.state.Head()
}

func (s *session) At(rev ident.Revision) (rinq.Revision, error) {
	return s.state.At(rev)
}

func (s *session) Attrs() (ident.Ref, attributes.Catalog) {
	return s.state.Attrs()
}

func (s *session) AttrsIn(ns string) (ident.Ref, attributes.VTable) {
	return s.state.AttrsIn(ns)
}

func (s *session) TryUpdate(ref ident.Ref, ns string, attrs attributes.List) (rinq.Revision, *attributes.Diff, error) {
	return s.state.TryUpdate(ref, ns, attrs)
}

func (s *session) TryClear(ref ident.Ref, ns string) (rinq.Revision, *attributes.Diff, error) {
	return s.state.TryClear(ref, ns)
}

func (s *session) TryDestroy(ref ident.Ref) error {
	return s.state.TryDestroy(ref)
}
