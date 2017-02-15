package internals

import "github.com/over-pass/overpass-go/src/overpass"

// SessionStore is an interface for retreiving sessions.
type SessionStore interface {
	Get(overpass.SessionID) (overpass.Session, error)
	Find(overpass.Constraint) ([]overpass.Session, error)
}
