package localsession

import (
	"sync"

	revisionpkg "github.com/rinq/rinq-go/src/internal/revision"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// Store is a collection of sessions and their state.
type Store interface {
	revisionpkg.Store

	Add(rinq.Session, *State)
	Remove(ident.SessionID)
	Get(ident.SessionID) (rinq.Session, *State, bool)
	Each(fn func(rinq.Session, *State))
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
	State   *State
}

func (s *store) Add(sess rinq.Session, state *State) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.entries[sess.ID()] = storeEntry{sess, state}
}

func (s *store) Remove(id ident.SessionID) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.entries, id)
}

func (s *store) Get(id ident.SessionID) (rinq.Session, *State, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	e, ok := s.entries[id]
	return e.Session, e.State, ok
}

func (s *store) Each(fn func(rinq.Session, *State)) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, e := range s.entries {
		fn(e.Session, e.State)
	}
}

func (s *store) GetRevision(ref ident.Ref) (rinq.Revision, error) {
	_, state, ok := s.Get(ref.ID)
	if ok {
		return state.At(ref.Rev)
	}

	return revisionpkg.Closed(ref), nil
}
