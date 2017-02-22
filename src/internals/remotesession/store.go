package remotesession

import (
	"sync"
	"time"

	"github.com/over-pass/overpass-go/src/internals/command"
	revisionpkg "github.com/over-pass/overpass-go/src/internals/revision"
	"github.com/over-pass/overpass-go/src/internals/service"
	"github.com/over-pass/overpass-go/src/overpass"
)

// Store is a local cache of remote revisions.
type Store interface {
	revisionpkg.Store
	service.Service
}

type store struct {
	service.Service
	closer *service.Closer

	peerID overpass.PeerID
	client *client
	ticker *time.Ticker
	logger overpass.Logger

	mutex sync.Mutex
	cache map[overpass.SessionID]*catalogCacheEntry
}

// NewStore returns a new store for revisions of remote sessions.
func NewStore(
	peerID overpass.PeerID,
	invoker command.Invoker,
	pruneInterval time.Duration,
	logger overpass.Logger,
) Store {
	svc, closer := service.NewImpl()

	s := &store{
		Service: svc,
		closer:  closer,

		peerID: peerID,
		client: &client{
			peerID:  peerID,
			invoker: invoker,
		},
		ticker: time.NewTicker(pruneInterval),
		logger: logger,
		cache:  map[overpass.SessionID]*catalogCacheEntry{},
	}

	go s.monitor()

	return s
}

type catalogCacheEntry struct {
	Catalog *catalog
	Marked  bool
}

func (s *store) GetRevision(ref overpass.SessionRef) (overpass.Revision, error) {
	cat := s.getCatalog(ref.ID)
	return cat.At(ref.Rev), nil
}

func (s *store) getCatalog(id overpass.SessionID) *catalog {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if entry, ok := s.cache[id]; ok {
		entry.Marked = false
		return entry.Catalog
	}

	if len(s.cache) == 0 && s.logger.IsDebug() {
		logCacheAdd(s.logger, s.peerID, id)
	}

	cat := &catalog{id: id, client: s.client}
	s.cache[id] = &catalogCacheEntry{cat, false}

	return cat
}

func (s *store) monitor() {
	for {
		select {
		case <-s.ticker.C:
			s.prune()
		case <-s.closer.Stop():
			s.ticker.Stop()
			s.closer.Close(nil)
		}
	}
}

func (s *store) prune() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for id, entry := range s.cache {
		if entry.Marked {
			delete(s.cache, id)

			if s.logger.IsDebug() {
				logCacheRemove(s.logger, s.peerID, id)
			}
		} else {
			entry.Marked = true

			if s.logger.IsDebug() {
				logCacheMark(s.logger, s.peerID, id)
			}
		}
	}
}
