package overpassamqp

import (
	"context"
	"log"
	"sync"

	"github.com/over-pass/overpass-go/src/internals"
	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/over-pass/overpass-go/src/overpass"
)

// localSession represents a session owned by a peer running in this process.
type localSession struct {
	id       overpass.SessionID
	invoker  internals.Invoker
	notifier internals.Notifier
	listener internals.Listener
	logger   *log.Logger

	done chan struct{}

	mutex      sync.RWMutex
	current    *localRevision
	messageSeq uint32
}

func newLocalSession(
	id overpass.SessionID,
	invoker internals.Invoker,
	notifier internals.Notifier,
	listener internals.Listener,
	logger *log.Logger,
) *localSession {
	sess := &localSession{
		id:       id,
		invoker:  invoker,
		notifier: notifier,
		listener: listener,
		logger:   logger,
		done:     make(chan struct{}),
	}
	sess.current = &localRevision{
		session: sess,
		ref:     id.At(0),
	}

	sess.logger.Printf(
		"%s session created",
		sess.current.Ref().ShortString(),
	)

	return sess
}

func (s *localSession) ID() overpass.SessionID {
	return s.id
}

func (s *localSession) CurrentRevision() (overpass.Revision, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.current == nil {
		return nil, overpass.NotFoundError{ID: s.id}
	}

	return s.current, nil
}

func (s *localSession) Call(ctx context.Context, ns, cmd string, p *overpass.Payload) (*overpass.Payload, error) {
	msgID, err := s.nextMessageID()
	if err != nil {
		return nil, err
	}

	return s.invoker.CallBalanced(ctx, msgID, ns, cmd, p)
}

func (s *localSession) Execute(ctx context.Context, ns, cmd string, p *overpass.Payload) error {
	msgID, err := s.nextMessageID()
	if err != nil {
		return err
	}

	return s.invoker.ExecuteBalanced(ctx, msgID, ns, cmd, p)
}

func (s *localSession) ExecuteMany(ctx context.Context, ns, cmd string, p *overpass.Payload) error {
	msgID, err := s.nextMessageID()
	if err != nil {
		return err
	}

	return s.invoker.ExecuteMulticast(ctx, msgID, ns, cmd, p)
}

func (s *localSession) Notify(ctx context.Context, target overpass.SessionID, typ string, p *overpass.Payload) error {
	msgID, err := s.nextMessageID()
	if err != nil {
		return err
	}

	return s.notifier.NotifyUnicast(ctx, msgID, target, typ, p)
}

func (s *localSession) NotifyMany(ctx context.Context, con overpass.Constraint, typ string, p *overpass.Payload) error {
	msgID, err := s.nextMessageID()
	if err != nil {
		return err
	}

	return s.notifier.NotifyMulticast(ctx, msgID, con, typ, p)
}

func (s *localSession) Listen(handler overpass.NotificationHandler) error {
	if handler == nil {
		panic("handler must not be nil")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.current == nil {
		return overpass.NotFoundError{ID: s.id}
	}

	return s.listener.Listen(s.id, handler)
}

func (s *localSession) Unlisten() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.current == nil {
		return overpass.NotFoundError{ID: s.id}
	}

	return s.listener.Unlisten(s.id)
}

func (s *localSession) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.current != nil {
		s.destroy(context.Background())
	}
}

func (s *localSession) Done() <-chan struct{} {
	return s.done
}

func (s *localSession) ApplyUpdate(ctx context.Context, next *localRevision) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.current == nil {
		return overpass.NotFoundError{ID: s.id}
	}

	expected := s.current.Ref()
	expected.Rev++

	if next.Ref() != expected {
		return overpass.StaleUpdateError{Ref: next.Ref()}
	}

	s.current = next
	s.messageSeq = 0

	if corrID := amqputil.GetCorrelationID(ctx); corrID != "" {
		s.logger.Printf(
			"%s session updated {%s} [%s]",
			s.current.Ref().ShortString(),
			s.current,
			corrID,
		)
	} else {
		s.logger.Printf(
			"%s session updated {%s}",
			s.current.Ref().ShortString(),
			s.current,
		)
	}

	return nil
}

func (s *localSession) ApplyClose(ctx context.Context, ref overpass.SessionRef) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.current == nil {
		return nil
	}

	if ref != s.current.Ref() {
		return overpass.StaleUpdateError{Ref: ref}
	}

	s.destroy(ctx)

	return nil
}

func (s *localSession) nextMessageID() (msgID overpass.MessageID, err error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.current == nil {
		err = overpass.NotFoundError{ID: s.id}
	} else {
		s.messageSeq++
		msgID = overpass.MessageID{
			Session: s.current.Ref(),
			Seq:     s.messageSeq,
		}
	}

	return
}

// destroy the session. It is assumed that s.mutex is already acquired for writing.
func (s *localSession) destroy(ctx context.Context) {
	ref := s.current.Ref()
	s.current = nil
	close(s.done)

	s.listener.Unlisten(s.id) // TODO: log error, probably

	if corrID := amqputil.GetCorrelationID(ctx); corrID != "" {
		s.logger.Printf(
			"%s session destroyed [%s]",
			ref.ShortString(),
			corrID,
		)
	} else {
		s.logger.Printf(
			"%s session destroyed",
			ref.ShortString(),
		)
	}
}
