package ident

import "fmt"

// Revision holds the "version" of a session. A session's revision is
// incremented when a change is made to its attribute table. A session that has
// never been modified, and hence has no attributes always has a revision of 0.
type Revision uint32

// Ref refers to a session at a specific revision.
type Ref struct {
	ID  SessionID
	Rev Revision
}

// Validate returns nil if the Ref is valid.
func (ref Ref) Validate() error {
	if ref.ID.Validate() == nil {
		return nil
	}

	return fmt.Errorf("session reference %s is invalid", ref)
}

// Before returns true if this ref's revision is before r.
func (ref Ref) Before(r Ref) bool {
	if ref.ID != r.ID {
		panic("can not compare references from different sessions")
	}

	return ref.Rev < r.Rev
}

// After returns true if this ref's revision is after r.
func (ref Ref) After(r Ref) bool {
	if ref.ID != r.ID {
		panic("can not compare references from different sessions")
	}

	return ref.Rev > r.Rev
}

// ShortString returns a string representation based on the session's short
// string representation.
func (ref Ref) ShortString() string {
	return fmt.Sprintf("%s@%d", ref.ID.ShortString(), ref.Rev)
}

func (ref Ref) String() string {
	return fmt.Sprintf("%s@%d", ref.ID, ref.Rev)
}
