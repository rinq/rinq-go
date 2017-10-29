package localsession

import (
	"errors"
	"sync"

	"github.com/rinq/rinq-go/src/internal/attributes"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// State represents a session's revisioned state.
type State struct {
	mutex  sync.RWMutex
	ref    ident.Ref
	attrs  attributes.Catalog
	seq    uint32
	done   chan struct{}
	logger rinq.Logger
}

// Ref returns the most recent session-ref.
// The ref's revision increments each time a call to TryUpdate() succeeds.
func (s *State) Ref() ident.Ref {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.ref
}

// NextMessageID generates a unique message ID from the current session-ref.
func (s *State) NextMessageID() (ident.MessageID, attributes.Catalog) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.seq++
	return s.ref.Message(s.seq), s.attrs
}

// Head returns the most recent revision.
// It is conceptually equivalent to s.At(s.Ref().Rev).
func (s *State) Head() rinq.Revision {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return &revision{s.ref, s, s.attrs, s.logger}
}

// At returns a revision representing the state at a specific revision
// number. The revision can not be newer than the current session-ref.
func (s *State) At(rev ident.Revision) (rinq.Revision, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.ref.Rev < rev {
		return nil, errors.New("revision is from the future")
	}

	return &revision{
		s.ref.ID.At(rev),
		s,
		s.attrs,
		s.logger,
	}, nil
}

// Attrs returns all attributes at the most recent revision.
func (s *State) Attrs() (ident.Ref, attributes.Catalog) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.ref, s.attrs
}

// AttrsIn returns all attributes in the ns namespace at the most recent revision.
func (s *State) AttrsIn(ns string) (ident.Ref, attributes.VTable) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.ref, s.attrs[ns]
}

// TryUpdate adds or updates attributes in the ns namespace of the attribute
// table and returns the new head revision.
//
// The operation fails if ref is not the current session-ref, attrs includes
// changes to frozen attributes, or the session has been destroyed.
func (s *State) TryUpdate(
	ref ident.Ref,
	ns string,
	attrs attributes.List,
) (rinq.Revision, *attributes.Diff, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	select {
	case <-s.done:
		return nil, nil, rinq.NotFoundError{ID: s.ref.ID}
	default:
	}

	if ref != s.ref {
		return nil, nil, rinq.StaleUpdateError{Ref: ref}
	}

	nextRev := ref.Rev + 1
	nextAttrs := s.attrs[ns].Clone()
	diff := attributes.NewDiff(ns, nextRev)

	for _, attr := range attrs {
		entry, exists := nextAttrs[attr.Key]

		if attr.Value == entry.Value && attr.IsFrozen == entry.IsFrozen {
			continue
		}

		if entry.IsFrozen {
			return nil, nil, rinq.FrozenAttributesError{Ref: ref}
		}

		entry.Attr = attr
		entry.UpdatedAt = nextRev
		if !exists {
			entry.CreatedAt = nextRev
		}

		nextAttrs[attr.Key] = entry
		diff.Append(entry)
	}

	s.ref.Rev = nextRev
	s.seq = 0

	if !diff.IsEmpty() {
		s.attrs = s.attrs.WithNamespace(ns, nextAttrs)
	}

	return &revision{
		s.ref,
		s,
		s.attrs,
		s.logger,
	}, diff, nil
}

// TryClear updates all attributes in the ns namespace of the attribute
// table to the empty string and returns the new head revision.
//
// The operation fails if ref is not the current session-ref, there are any
// frozen attributes, or the session has been destroyed.
func (s *State) TryClear(
	ref ident.Ref,
	ns string,
) (rinq.Revision, *attributes.Diff, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	select {
	case <-s.done:
		return nil, nil, rinq.NotFoundError{ID: s.ref.ID}
	default:
	}

	if ref != s.ref {
		return nil, nil, rinq.StaleUpdateError{Ref: ref}
	}

	attrs := s.attrs[ns]
	nextRev := ref.Rev + 1
	nextAttrs := attributes.VTable{}
	diff := attributes.NewDiff(ns, nextRev)

	for _, entry := range attrs {
		if entry.Value != "" {
			if entry.IsFrozen {
				return nil, nil, rinq.FrozenAttributesError{Ref: ref}
			}

			entry.Value = ""
			entry.UpdatedAt = nextRev
			diff.Append(entry)
		}

		nextAttrs[entry.Key] = entry
	}

	s.ref.Rev = nextRev
	s.seq = 0

	if !diff.IsEmpty() {
		s.attrs = s.attrs.WithNamespace(ns, nextAttrs)
	}

	return &revision{
		s.ref,
		s,
		s.attrs,
		s.logger,
	}, diff, nil
}

// TryDestroy destroys the session, preventing further updates.
//
// The operation fails if ref is not the current session-ref. It is not an
// error to destroy an already-destroyed session.
func (s *State) TryDestroy(ref ident.Ref) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if ref != s.ref {
		return rinq.StaleUpdateError{Ref: ref}
	}

	select {
	case <-s.done:
	default:
		close(s.done)
	}

	return nil
}

// Close forcefully destroys the session, preventing further updates.
// It is not an error to close an already-destroyed session.
func (s *State) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	select {
	case <-s.done:
	default:
		close(s.done)
	}
}

// Done returns a channel that is closed when the session is destroyed.
func (s *State) Done() <-chan struct{} {
	return s.done
}
