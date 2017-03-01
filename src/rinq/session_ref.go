package rinq

import "fmt"

// RevisionNumber holds the "version" of a session. A session's revision is
// incremented when a change is made to its attribute table. A session that has
// never been modified, and hence has no attributes always has a revision of 0.
type RevisionNumber uint32

// SessionRef refers to a session at a specific revision.
type SessionRef struct {
	ID  SessionID
	Rev RevisionNumber
}

// Validate returns nil if the Ref is valid.
func (ref SessionRef) Validate() error {
	if ref.ID.Validate() == nil {
		return nil
	}

	return fmt.Errorf("session reference %s is invalid", ref)
}

// Before returns true if this ref's revision is before r.
func (ref SessionRef) Before(r SessionRef) bool {
	if ref.ID != r.ID {
		panic("can not compare references from different sessions")
	}

	return ref.Rev < r.Rev
}

// After returns true if this ref's revision is after r.
func (ref SessionRef) After(r SessionRef) bool {
	if ref.ID != r.ID {
		panic("can not compare references from different sessions")
	}

	return ref.Rev > r.Rev
}

// ShortString returns a string representation based on the session's short
// string representation.
func (ref SessionRef) ShortString() string {
	return fmt.Sprintf("%s@%d", ref.ID.ShortString(), ref.Rev)
}

func (ref SessionRef) String() string {
	return fmt.Sprintf("%s@%d", ref.ID, ref.Rev)
}
