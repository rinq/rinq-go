package localsession

import (
	"context"
	"sync"
	"time"

	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/over-pass/overpass-go/src/internals/attrmeta"
	"github.com/over-pass/overpass-go/src/internals/bufferpool"
	"github.com/over-pass/overpass-go/src/internals/command"
	"github.com/over-pass/overpass-go/src/internals/deferutil"
	"github.com/over-pass/overpass-go/src/internals/notify"
	"github.com/over-pass/overpass-go/src/overpass"
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
		sess.Close()
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

func (s *session) Call(ctx context.Context, ns, cmd string, p *overpass.Payload) (*overpass.Payload, error) {
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
	corrID, payload, err := s.invoker.CallBalanced(ctx, msgID, ns, cmd, p)
	elapsed := time.Now().Sub(start) / time.Millisecond

	if err == context.DeadlineExceeded || err == context.Canceled {
		s.logger.Log(
			"%s called '%s::%s' command: %s (%dms, %d/o -/i) [%s]",
			msgID.ShortString(),
			ns,
			cmd,
			err,
			elapsed,
			p.Len(),
			corrID,
		)
	} else {
		switch e := err.(type) {
		case nil:
			s.logger.Log(
				"%s called '%s::%s' command successfully (%dms, %d/o %d/i) [%s]",
				msgID.ShortString(),
				ns,
				cmd,
				elapsed,
				p.Len(),
				payload.Len(),
				corrID,
			)
		case overpass.Failure:
			s.logger.Log(
				"%s called '%s::%s' command: '%s' failure (%dms, %d/o %d/i) [%s]",
				msgID.ShortString(),
				ns,
				cmd,
				e.Type,
				elapsed,
				p.Len(),
				payload.Len(),
				corrID,
			)
		case overpass.UnexpectedError:
			s.logger.Log(
				"%s called '%s::%s' command: '%s' error (%dms, %d/o 0/i) [%s]",
				msgID.ShortString(),
				ns,
				cmd,
				e,
				elapsed,
				p.Len(),
				corrID,
			)
		}
	}

	return payload, err
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
	corrID, err := s.invoker.ExecuteBalanced(ctx, msgID, ns, cmd, p)

	if err == nil {
		s.logger.Log(
			"%s executed '%s::%s' command (%d/o) [%s]",
			msgID.ShortString(),
			ns,
			cmd,
			p.Len(),
			corrID,
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
	corrID, err := s.invoker.ExecuteMulticast(ctx, msgID, ns, cmd, p)

	if err == nil {
		s.logger.Log(
			"%s executed '%s::%s' command on multiple peers (%d/o) [%s]",
			msgID.ShortString(),
			ns,
			cmd,
			p.Len(),
			corrID,
		)
	}

	return err
}

func (s *session) Notify(ctx context.Context, target overpass.SessionID, typ string, p *overpass.Payload) error {
	select {
	case <-s.done:
		return overpass.NotFoundError{ID: s.id}
	default:
	}

	msgID := s.catalog.NextMessageID()
	corrID, err := s.notifier.NotifyUnicast(ctx, msgID, target, typ, p)

	if err == nil {
		s.logger.Log(
			"%s sent '%s' notification to %s (%d/o) [%s]",
			msgID.ShortString(),
			typ,
			target.ShortString(),
			p.Len(),
			corrID,
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
	corrID, err := s.notifier.NotifyMulticast(ctx, msgID, con, typ, p)

	if err == nil {
		s.logger.Log(
			"%s sent '%s' notification to sessions matching {%s} (%d/o) [%s]",
			msgID.ShortString(),
			typ,
			con,
			p.Len(),
			corrID,
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
				amqputil.GetCorrelationID(ctx),
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
	unlock := deferutil.Lock(&s.mutex)
	defer unlock()

	select {
	case <-s.done:
		return
	default:
	}

	close(s.done)
	s.catalog.Close()
	s.listener.Unlisten(s.id)

	unlock()

	ref, attrs := s.catalog.Attrs()

	buffer := bufferpool.Get()
	defer bufferpool.Put(buffer)

	for _, attr := range attrs {
		if !attr.IsFrozen && attr.Value == "" {
			continue
		}

		if buffer.Len() != 0 {
			buffer.WriteString(", ")
		}

		attrmeta.Write(buffer, attr)
	}

	s.logger.Log(
		"%s session destroyed {%s}",
		ref.ShortString(),
		buffer,
	)
}

func (s *session) Done() <-chan struct{} {
	return s.done
}
