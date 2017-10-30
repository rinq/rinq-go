package localsession

import (
	"errors"

	"github.com/rinq/rinq-go/src/internal/attributes"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// This file contains methods of the Session struct that are not defined in
// rinq.Session. See session.go for the implementation of rinq.Session.

// At returns a revision representing the state at a specific revision
// number. The revision can not be newer than the current session-ref.
func (s *Session) At(rev ident.Revision) (rinq.Revision, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.ref.Rev < rev {
		return nil, errors.New("revision is from the future")
	}

	return &revision{
		s,
		s.ref,
		s.attrs,
		s.logger,
	}, nil
}

// Attrs returns all attributes at the most recent revision.
func (s *Session) Attrs() (ident.Ref, attributes.Catalog) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.ref, s.attrs
}

// AttrsIn returns all attributes in the ns namespace at the most recent revision.
func (s *Session) AttrsIn(ns string) (ident.Ref, attributes.VTable) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.ref, s.attrs[ns]
}

// TryUpdate adds or updates attributes in the ns namespace of the attribute
// table and returns the new head revision.
//
// The operation fails if ref is not the current session-ref, attrs includes
// changes to frozen attributes, or the session has been destroyed.
func (s *Session) TryUpdate(rev ident.Revision, ns string, attrs attributes.List) (rinq.Revision, *attributes.Diff, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isDestroyed {
		return nil, nil, rinq.NotFoundError{ID: s.ref.ID}
	}

	if rev != s.ref.Rev {
		return nil, nil, rinq.StaleUpdateError{Ref: s.ref.ID.At(rev)}
	}

	nextRev := rev + 1
	nextAttrs := s.attrs[ns].Clone()
	diff := attributes.NewDiff(ns, nextRev)

	for _, attr := range attrs {
		entry, exists := nextAttrs[attr.Key]

		if attr.Value == entry.Value && attr.IsFrozen == entry.IsFrozen {
			continue
		}

		if entry.IsFrozen {
			return nil, nil, rinq.FrozenAttributesError{Ref: s.ref.ID.At(rev)}
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
	s.msgSeq = 0

	if !diff.IsEmpty() {
		s.attrs = s.attrs.WithNamespace(ns, nextAttrs)
	}

	return &revision{
		s,
		s.ref,
		s.attrs,
		s.logger,
	}, diff, nil
}

// TryClear updates all attributes in the ns namespace of the attribute
// table to the empty string and returns the new head revision.
//
// The operation fails if ref is not the current session-ref, there are any
// frozen attributes, or the session has been destroyed.
func (s *Session) TryClear(rev ident.Revision, ns string) (rinq.Revision, *attributes.Diff, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isDestroyed {
		return nil, nil, rinq.NotFoundError{ID: s.ref.ID}
	}

	if rev != s.ref.Rev {
		return nil, nil, rinq.StaleUpdateError{Ref: s.ref.ID.At(rev)}
	}

	attrs := s.attrs[ns]
	nextRev := rev + 1
	nextAttrs := attributes.VTable{}
	diff := attributes.NewDiff(ns, nextRev)

	for _, entry := range attrs {
		if entry.Value != "" {
			if entry.IsFrozen {
				return nil, nil, rinq.FrozenAttributesError{Ref: s.ref.ID.At(rev)}
			}

			entry.Value = ""
			entry.UpdatedAt = nextRev
			diff.Append(entry)
		}

		nextAttrs[entry.Key] = entry
	}

	s.ref.Rev = nextRev
	s.msgSeq = 0

	if !diff.IsEmpty() {
		s.attrs = s.attrs.WithNamespace(ns, nextAttrs)
	}

	return &revision{
		s,
		s.ref,
		s.attrs,
		s.logger,
	}, diff, nil
}

// TryDestroy destroys the session, preventing further updates.
//
// The operation fails if ref is not the current session-ref. It is not an
// error to destroy an already-destroyed session.
//
// first is true if this call caused the session to be destroyed.
func (s *Session) TryDestroy(rev ident.Revision) (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if rev != s.ref.Rev {
		return false, rinq.StaleUpdateError{Ref: s.ref.ID.At(rev)}
	}

	if s.isDestroyed {
		return false, nil
	}

	s.destroy()

	return true, nil
}

// destroy marks the session as destroyed removes any callbacks registered with
// the command and notification subsystems.
func (s *Session) destroy() {
	s.isDestroyed = true

	s.invoker.SetAsyncHandler(s.ref.ID, nil)
	_ = s.listener.UnlistenAll(s.ref.ID)

	go func() {
		// close the done channel only after all pending calls have finished
		s.calls.Wait()
		close(s.done)
	}()
}

// nextMessageID returns a new unique message ID generated from the current
// session-ref, and the attributes as they existed at that ref.
func (s *Session) nextMessageID() ident.MessageID {
	s.msgSeq++
	return s.ref.Message(s.msgSeq)
}
