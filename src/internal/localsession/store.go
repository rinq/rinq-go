package localsession

import (
	"sync"

	"github.com/rinq/rinq-go/src/internal/revisions"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// Store is a collection of local sessions which provides an implementation
// of revisions.Store.
type Store struct {
	mutex    sync.RWMutex
	sessions map[ident.SessionID]*Session
}

// NewStore returns a new session store.
func NewStore() *Store {
	return &Store{
		sessions: map[ident.SessionID]*Session{},
	}
}

// Add adds a session to the store.
func (s *Store) Add(sess *Session) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.sessions[sess.ID()] = sess
}

// Remove removes a session to from the store.
func (s *Store) Remove(id ident.SessionID) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.sessions, id)
}

// Get fetches a session from the store by its ID.
func (s *Store) Get(id ident.SessionID) (sess *Session, ok bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	sess, ok = s.sessions[id]
	return
}

// Each calls fn(sess) for each session in the store.
func (s *Store) Each(fn func(*Session)) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, sess := range s.sessions {
		fn(sess)
	}
}

// GetRevision returns the session revision for the given ref.
func (s *Store) GetRevision(ref ident.Ref) (rinq.Revision, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if sess, ok := s.sessions[ref.ID]; ok {
		return sess.At(ref.Rev)
	}

	return revisions.Closed(ref.ID), nil
}
