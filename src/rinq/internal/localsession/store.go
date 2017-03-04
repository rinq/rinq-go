package localsession

import (
	"sync"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	revisionpkg "github.com/rinq/rinq-go/src/rinq/internal/revision"
)

// Store is a collection of sessions and their catalogs.
type Store interface {
	revisionpkg.Store

	Add(rinq.Session, Catalog)
	Remove(ident.SessionID)
	Get(ident.SessionID) (rinq.Session, Catalog, bool)
	Each(fn func(rinq.Session, Catalog))
}

type store struct {
	mutex   sync.RWMutex
	entries map[ident.SessionID]storeEntry
}

// NewStore returns a new session store.
func NewStore() Store {
	return &store{
		entries: map[ident.SessionID]storeEntry{},
	}
}

type storeEntry struct {
	Session rinq.Session
	Catalog Catalog
}

func (s *store) Add(sess rinq.Session, cat Catalog) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.entries[sess.ID()] = storeEntry{sess, cat}
}

func (s *store) Remove(id ident.SessionID) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.entries, id)
}

func (s *store) Get(id ident.SessionID) (rinq.Session, Catalog, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	e, ok := s.entries[id]
	return e.Session, e.Catalog, ok
}

func (s *store) Each(fn func(rinq.Session, Catalog)) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, e := range s.entries {
		fn(e.Session, e.Catalog)
	}
}

func (s *store) GetRevision(ref ident.Ref) (rinq.Revision, error) {
	_, cat, ok := s.Get(ref.ID)
	if ok {
		return cat.At(ref.Rev)
	}

	return revisionpkg.Closed(ref), nil
}
