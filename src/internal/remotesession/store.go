package remotesession

import (
	"sync"
	"time"

	"github.com/jmalloc/twelf/src/twelf"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/internal/command"
	"github.com/rinq/rinq-go/src/internal/revisions"
	"github.com/rinq/rinq-go/src/internal/service"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// Store is a local cache of remote revisions.
type Store interface {
	revisions.Store
	service.Service
}

type store struct {
	service.Service
	sm *service.StateMachine

	peerID   ident.PeerID
	client   *client
	interval time.Duration
	logger   twelf.Logger

	mutex sync.Mutex
	cache map[ident.SessionID]*cacheEntry
}

// NewStore returns a new store for revisions of remote sessions.
func NewStore(
	peerID ident.PeerID,
	invoker command.Invoker,
	pruneInterval time.Duration,
	logger twelf.Logger,
	tracer opentracing.Tracer,
) Store {
	s := &store{
		peerID:   peerID,
		client:   newClient(peerID, invoker, logger, tracer),
		interval: pruneInterval,
		logger:   logger,
		cache:    map[ident.SessionID]*cacheEntry{},
	}

	s.sm = service.NewStateMachine(s.run, nil)
	s.Service = s.sm

	go s.sm.Run()

	return s
}

type cacheEntry struct {
	Session *session
	Marked  bool
}

func (s *store) GetRevision(ref ident.Ref) (rinq.Revision, error) {
	sess := s.getSession(ref.ID)
	return sess.At(ref.Rev), nil
}

func (s *store) getSession(id ident.SessionID) *session {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if entry, ok := s.cache[id]; ok {
		entry.Marked = false
		return entry.Session
	}

	sess := newSession(id, s.client)
	s.cache[id] = &cacheEntry{sess, false}
	logCacheAdd(s.logger, s.peerID, id)

	return sess
}

func (s *store) run() (service.State, error) {
	for {
		select {
		case <-time.After(s.interval):
			s.prune()

		case <-s.sm.Graceful:
			return nil, nil

		case <-s.sm.Forceful:
			return nil, nil
		}
	}
}

func (s *store) prune() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for id, entry := range s.cache {
		if entry.Marked {
			delete(s.cache, id)
			logCacheRemove(s.logger, s.peerID, id)
		} else {
			entry.Marked = true
			logCacheMark(s.logger, s.peerID, id)
		}
	}
}
