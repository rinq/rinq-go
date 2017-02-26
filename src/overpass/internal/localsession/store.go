package localsession

import (
	"sync"

	revisionpkg "github.com/over-pass/overpass-go/src/overpass/internal/revision"
	"github.com/over-pass/overpass-go/src/overpass"
)

// Store is a collection of sessions and their catalogs.
type Store interface {
	revisionpkg.Store

	Add(overpass.Session, Catalog)
	Remove(overpass.SessionID)
	Get(overpass.SessionID) (overpass.Session, Catalog, bool)
	Each(fn func(overpass.Session, Catalog))
}

type store struct {
	mutex   sync.RWMutex
	entries map[overpass.SessionID]storeEntry
}

// NewStore returns a new session store.
func NewStore() Store {
	return &store{
		entries: map[overpass.SessionID]storeEntry{},
	}
}

type storeEntry struct {
	Session overpass.Session
	Catalog Catalog
}

func (s *store) Add(sess overpass.Session, cat Catalog) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.entries[sess.ID()] = storeEntry{sess, cat}
}

func (s *store) Remove(id overpass.SessionID) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.entries, id)
}

func (s *store) Get(id overpass.SessionID) (overpass.Session, Catalog, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	e, ok := s.entries[id]
	return e.Session, e.Catalog, ok
}

func (s *store) Each(fn func(overpass.Session, Catalog)) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, e := range s.entries {
		fn(e.Session, e.Catalog)
	}
}

func (s *store) GetRevision(ref overpass.SessionRef) (overpass.Revision, error) {
	_, cat, ok := s.Get(ref.ID)
	if ok {
		return cat.At(ref.Rev)
	}

	return revisionpkg.Closed(ref), nil
}
