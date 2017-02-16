package localsession

import (
	"context"
	"log"
	"sync"

	"github.com/over-pass/overpass-go/src/internals"
	"github.com/over-pass/overpass-go/src/overpass"
)

type session struct {
	id       overpass.SessionID
	catalog  Catalog
	invoker  internals.Invoker
	notifier internals.Notifier
	listener internals.Listener
	logger   *log.Logger
	done     chan struct{}

	// mutex guards Listen(), Unlisten() and Close() so that we can guarantee
	// that the call to listener.Unlisten() in Close() is actually final.
	mutex sync.RWMutex
}

// NewSession returns a new local session.
func NewSession(
	id overpass.SessionID,
	catalog Catalog,
	invoker internals.Invoker,
	notifier internals.Notifier,
	listener internals.Listener,
	logger *log.Logger,
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

	sess.logger.Printf(
		"%s session created",
		sess.catalog.Ref().ShortString(),
	)

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
	select {
	case <-s.done:
		return nil, overpass.NotFoundError{ID: s.id}
	default:
		return s.invoker.CallBalanced(ctx, s.catalog.NextMessageID(), ns, cmd, p)
	}
}

func (s *session) Execute(ctx context.Context, ns, cmd string, p *overpass.Payload) error {
	select {
	case <-s.done:
		return overpass.NotFoundError{ID: s.id}
	default:
		return s.invoker.ExecuteBalanced(ctx, s.catalog.NextMessageID(), ns, cmd, p)
	}
}

func (s *session) ExecuteMany(ctx context.Context, ns, cmd string, p *overpass.Payload) error {
	select {
	case <-s.done:
		return overpass.NotFoundError{ID: s.id}
	default:
		return s.invoker.ExecuteMulticast(ctx, s.catalog.NextMessageID(), ns, cmd, p)
	}
}

func (s *session) Notify(ctx context.Context, target overpass.SessionID, typ string, p *overpass.Payload) error {
	select {
	case <-s.done:
		return overpass.NotFoundError{ID: s.id}
	default:
		return s.notifier.NotifyUnicast(ctx, s.catalog.NextMessageID(), target, typ, p)
	}

}

func (s *session) NotifyMany(ctx context.Context, con overpass.Constraint, typ string, p *overpass.Payload) error {
	select {
	case <-s.done:
		return overpass.NotFoundError{ID: s.id}
	default:
		return s.notifier.NotifyMulticast(ctx, s.catalog.NextMessageID(), con, typ, p)
	}
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
		return s.listener.Listen(s.id, handler)
	}
}

func (s *session) Unlisten() error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	select {
	case <-s.done:
		return overpass.NotFoundError{ID: s.id}
	default:
		return s.listener.Unlisten(s.id)
	}
}

func (s *session) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	select {
	case <-s.done:
		return
	default:
	}

	close(s.done)
	s.catalog.Close()
	s.listener.Unlisten(s.id)

	s.logger.Printf(
		"%s session destroyed", // TODO: log args
		s.catalog.Ref().ShortString(),
	)
}

func (s *session) Done() <-chan struct{} {
	return s.done
}
