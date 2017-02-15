package overpassamqp

import (
	"sync"

	"github.com/over-pass/overpass-go/src/internals"
	"github.com/over-pass/overpass-go/src/internals/deferutil"
	"github.com/over-pass/overpass-go/src/overpass"
)

// localStore is a thread-safe collection of local sessions that implements
// services.SessionRepository and services.RevisionRepository.
type localStore struct {
	mutex    sync.RWMutex
	sessions map[overpass.SessionID]*localSession
}

func (s *localStore) Add(sess *localSession) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.sessions == nil {
		s.sessions = map[overpass.SessionID]*localSession{}
	}

	s.sessions[sess.ID()] = sess
}

func (s *localStore) Remove(sess *localSession) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.sessions, sess.ID())
}

func (s *localStore) Get(id overpass.SessionID) (overpass.Session, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if sess, ok := s.sessions[id]; ok {
		return sess, nil
	}

	return nil, overpass.NotFoundError{ID: id}
}

func (s *localStore) Find(
	constraint overpass.Constraint,
) ([]overpass.Session, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var sessions []overpass.Session

	for _, sess := range s.sessions {
		rev, err := sess.CurrentRevision()
		if overpass.IsNotFound(err) {
			continue
		} else if err != nil {
			return nil, err
		}

		if rev.(*localRevision).Match(constraint) {
			sessions = append(sessions, sess)
		}
	}

	return sessions, nil
}

func (s *localStore) GetRevision(ref overpass.SessionRef) (overpass.Revision, error) {
	var sess *localSession
	deferutil.RWith(&s.mutex, func() {
		sess = s.sessions[ref.ID]
	})

	if sess == nil {
		return internals.NewClosedRevision(ref), nil
	}

	rev, err := sess.CurrentRevision()
	if overpass.IsNotFound(err) {
		return internals.NewClosedRevision(ref), nil
	} else if err != nil {
		return nil, err
	}

	return rev.(*localRevision).At(ref)
}
