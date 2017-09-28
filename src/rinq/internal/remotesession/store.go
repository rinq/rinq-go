package remotesession

import (
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/command"
	revisionpkg "github.com/rinq/rinq-go/src/rinq/internal/revision"
	"github.com/rinq/rinq-go/src/rinq/internal/service"
)

// Store is a local cache of remote revisions.
type Store interface {
	revisionpkg.Store
	service.Service
}

type store struct {
	service.Service
	sm *service.StateMachine

	peerID   ident.PeerID
	client   *client
	interval time.Duration
	logger   rinq.Logger

	mutex sync.Mutex
	cache map[ident.SessionID]*catalogCacheEntry
}

// NewStore returns a new store for revisions of remote sessions.
func NewStore(
	peerID ident.PeerID,
	invoker command.Invoker,
	pruneInterval time.Duration,
	logger rinq.Logger,
	tracer opentracing.Tracer,
) Store {
	s := &store{
		peerID:   peerID,
		client:   newClient(peerID, invoker, logger, tracer),
		interval: pruneInterval,
		logger:   logger,
		cache:    map[ident.SessionID]*catalogCacheEntry{},
	}

	s.sm = service.NewStateMachine(s.run, nil)
	s.Service = s.sm

	go s.sm.Run()

	return s
}

type catalogCacheEntry struct {
	Catalog *catalog
	Marked  bool
}

func (s *store) GetRevision(ref ident.Ref) (rinq.Revision, error) {
	cat := s.getCatalog(ref.ID)
	return cat.At(ref.Rev), nil
}

func (s *store) getCatalog(id ident.SessionID) *catalog {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if entry, ok := s.cache[id]; ok {
		entry.Marked = false
		return entry.Catalog
	}

	cat := newCatalog(id, s.client)
	s.cache[id] = &catalogCacheEntry{cat, false}
	logCacheAdd(s.logger, s.peerID, id)

	return cat
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
