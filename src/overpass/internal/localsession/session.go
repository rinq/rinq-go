package localsession

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpass/internal/command"
	"github.com/over-pass/overpass-go/src/overpass/internal/notify"
	"github.com/over-pass/overpass-go/src/overpass/internal/trace"
)

type session struct {
	id       overpass.SessionID
	catalog  Catalog
	invoker  command.Invoker
	notifier notify.Notifier
	listener notify.Listener
	logger   overpass.Logger
	done     chan struct{}

	// mutex guards Call(), Listen(), Unlisten() and Close() so that Close()
	// waits for pending calls to complete or timeout, and to ensure that it's
	// call to listener.Unlisten() is not "undone" by the user.
	mutex sync.RWMutex
}

// NewSession returns a new local session.
func NewSession(
	id overpass.SessionID,
	catalog Catalog,
	invoker command.Invoker,
	notifier notify.Notifier,
	listener notify.Listener,
	logger overpass.Logger,
) overpass.Session {
	sess := &session{
		id:       id,
		catalog:  catalog,
		invoker:  invoker,
		notifier: notifier,
		logger:   logger,
		listener: listener,
		done:     make(chan struct{}),
	}

	sess.logger.Log(
		"%s session created",
		sess.catalog.Ref().ShortString(),
	)

	go func() {
		<-catalog.Done()
		sess.close()
	}()

	return sess
}

func (s *session) ID() overpass.SessionID {
	return s.id
}

func (s *session) CurrentRevision() (overpass.Revision, error) {
	select {
	case <-s.done:
		return nil, overpass.NotFoundError{ID: s.id}
	default:
		return s.catalog.Head(), nil
	}
}

func (s *session) Call(ctx context.Context, ns, cmd string, out *overpass.Payload) (*overpass.Payload, error) {
	if err := overpass.ValidateNamespace(ns); err != nil {
		return nil, err
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	select {
	case <-s.done:
		return nil, overpass.NotFoundError{ID: s.id}
	default:
	}

	msgID := s.catalog.NextMessageID()

	start := time.Now()
	traceID, in, err := s.invoker.CallBalanced(ctx, msgID, ns, cmd, out)
	elapsed := time.Now().Sub(start) / time.Millisecond

	logCall(s.logger, msgID, ns, cmd, elapsed, out, in, err, traceID)

	return in, err
}

func (s *session) CallAsync(ctx context.Context, ns, cmd string, out *overpass.Payload) (overpass.MessageID, error) {
	var msgID overpass.MessageID

	if err := overpass.ValidateNamespace(ns); err != nil {
		return msgID, err
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	select {
	case <-s.done:
		return msgID, overpass.NotFoundError{ID: s.id}
	default:
	}

	msgID = s.catalog.NextMessageID()

	traceID, err := s.invoker.CallBalancedAsync(ctx, msgID, ns, cmd, out)
	if err != nil {
		return msgID, err
	}

	logAsyncRequest(s.logger, msgID, ns, cmd, out, traceID)

	return msgID, nil
}

// SetAsyncHandler sets the asynchronous call handler.
//
// h is invoked for each command response received to a command request made
// with CallAsync().
func (s *session) SetAsyncHandler(h overpass.AsyncHandler) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	select {
	case <-s.done:
		return overpass.NotFoundError{ID: s.id}
	default:
	}

	s.invoker.SetAsyncHandler(
		s.id,
		func(
			ctx context.Context,
			msgID overpass.MessageID,
			ns string,
			cmd string,
			in *overpass.Payload,
			err error,
		) {
			logAsyncResponse(ctx, s.logger, msgID, ns, cmd, in, err)
			h(ctx, msgID, ns, cmd, in, err)
		},
	)

	return nil
}

func (s *session) Execute(ctx context.Context, ns, cmd string, p *overpass.Payload) error {
	if err := overpass.ValidateNamespace(ns); err != nil {
		return err
	}

	select {
	case <-s.done:
		return overpass.NotFoundError{ID: s.id}
	default:
	}

	msgID := s.catalog.NextMessageID()
	traceID, err := s.invoker.ExecuteBalanced(ctx, msgID, ns, cmd, p)

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

func (s *session) ExecuteMany(ctx context.Context, ns, cmd string, p *overpass.Payload) error {
	if err := overpass.ValidateNamespace(ns); err != nil {
		return err
	}

	select {
	case <-s.done:
		return overpass.NotFoundError{ID: s.id}
	default:
	}

	msgID := s.catalog.NextMessageID()
	traceID, err := s.invoker.ExecuteMulticast(ctx, msgID, ns, cmd, p)

	if err == nil {
		s.logger.Log(
			"%s executed '%s::%s' command on multiple peers (%d/o) [%s]",
			msgID.ShortString(),
			ns,
			cmd,
			p.Len(),
			traceID,
		)
	}

	return err
}

func (s *session) Notify(ctx context.Context, target overpass.SessionID, typ string, p *overpass.Payload) error {
	if err := target.Validate(); err != nil || target.Seq == 0 {
		return fmt.Errorf("session ID %s is invalid", target)
	}

	select {
	case <-s.done:
		return overpass.NotFoundError{ID: s.id}
	default:
	}

	msgID := s.catalog.NextMessageID()
	traceID, err := s.notifier.NotifyUnicast(ctx, msgID, target, typ, p)

	if err == nil {
		s.logger.Log(
			"%s sent '%s' notification to %s (%d/o) [%s]",
			msgID.ShortString(),
			typ,
			target.ShortString(),
			p.Len(),
			traceID,
		)
	}

	return err
}

func (s *session) NotifyMany(ctx context.Context, con overpass.Constraint, typ string, p *overpass.Payload) error {
	select {
	case <-s.done:
		return overpass.NotFoundError{ID: s.id}
	default:
	}

	msgID := s.catalog.NextMessageID()
	traceID, err := s.notifier.NotifyMulticast(ctx, msgID, con, typ, p)

	if err == nil {
		s.logger.Log(
			"%s sent '%s' notification to sessions matching {%s} (%d/o) [%s]",
			msgID.ShortString(),
			typ,
			con,
			p.Len(),
			traceID,
		)
	}

	return err
}

func (s *session) Listen(handler overpass.NotificationHandler) error {
	if handler == nil {
		panic("handler must not be nil")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	select {
	case <-s.done:
		return overpass.NotFoundError{ID: s.id}
	default:
	}

	changed, err := s.listener.Listen(
		s.id,
		func(
			ctx context.Context,
			target overpass.Session,
			n overpass.Notification,
		) {
			rev := s.catalog.Head()

			s.logger.Log(
				"%s received '%s' notification from %s (%d/i) [%s]",
				rev.Ref().ShortString(),
				n.Type,
				n.Source.Ref().ShortString(),
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
			"%s started listening for notifications",
			s.catalog.Ref().ShortString(),
		)
	}

	return nil
}

func (s *session) Unlisten() error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	select {
	case <-s.done:
		return overpass.NotFoundError{ID: s.id}
	default:
	}

	changed, err := s.listener.Unlisten(s.id)

	if err != nil {
		return err
	} else if changed && s.logger.IsDebug() {
		s.logger.Log(
			"%s stopped listening for notifications",
			s.catalog.Ref().ShortString(),
		)
	}

	return nil
}

func (s *session) Close() {
	if s.close() {
		logSessionClose(s.logger, s.catalog, "")
	}
}

func (s *session) close() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	select {
	case <-s.done:
		return false
	default:
		close(s.done)
		s.catalog.Close()
		s.invoker.SetAsyncHandler(s.id, nil)
		s.listener.Unlisten(s.id)
		return true
	}
}

func (s *session) Done() <-chan struct{} {
	return s.done
}
