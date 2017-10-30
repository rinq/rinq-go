package localsession

import (
	"sync"

	"github.com/rinq/rinq-go/src/internal/revisions"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// Store is a collection of sessions and their state.
type Store interface {
	revisions.Store

	Add(Session)
	Remove(ident.SessionID)
	Get(ident.SessionID) (Session, bool)
	Each(fn func(Session))
}

type store struct {
	mutex   sync.RWMutex
	entries map[ident.SessionID]Session
}

// NewStore returns a new session store.
func NewStore() Store {
	return &store{
		entries: map[ident.SessionID]Session{},
	}
}

func (s *store) Add(sess Session) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.entries[sess.ID()] = sess
}

func (s *store) Remove(id ident.SessionID) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.entries, id)
}

func (s *store) Get(id ident.SessionID) (Session, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	sess, ok := s.entries[id]
	return sess, ok
}

func (s *store) Each(fn func(Session)) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, sess := range s.entries {
		fn(sess)
	}
}

func (s *store) GetRevision(ref ident.Ref) (rinq.Revision, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if sess, ok := s.entries[ref.ID]; ok {
		return sess.At(ref.Rev)
	}

	return revisions.Closed(ref.ID), nil
}
