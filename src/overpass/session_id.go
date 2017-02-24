package overpass

import (
	"fmt"
	"regexp"
	"strconv"
)

// SessionID uniquely identifies a session within a network.
//
// Session IDs contain a peer component, and a 32-but sequence component.
// They are rendereds as a peer ID, followed by a period, then the sequence
// component as a decimal, such as "58AEE146-191C.45".
//
// Because the peer ID is embedded, the same uniqueness guarantees apply to the
// session ID as to the peer ID.
type SessionID struct {
	// Peer is the ID of the peer that owns the session.
	Peer PeerID

	// Seq is a monotonically increasing sequence allocated to each session in
	// the order it is created by the owning peer. Application sessions begin
	// with a sequence value of 1. The sequnce value zero is reserved for the
	// "zero-session", which is used internally by Overpass.
	Seq uint32
}

// ParseSessionID parses a string representation of a session ID.
func ParseSessionID(str string) (id SessionID, err error) {
	matches := sessionIDPattern.FindStringSubmatch(str)

	if len(matches) != 0 {
		// Read the peer ID clock component ...
		var value uint64
		value, err = strconv.ParseUint(matches[1], 16, 64)
		if err != nil {
			return
		}
		id.Peer.Clock = value

		// Read the peer ID random component ...
		value, err = strconv.ParseUint(matches[2], 16, 16)
		if err != nil {
			return
		}
		id.Peer.Rand = uint16(value)

		// Read the session ID sequence component ...
		value, err = strconv.ParseUint(matches[3], 10, 32)
		if err != nil {
			return
		}
		id.Seq = uint32(value)
	}

	err = id.Validate()
	return
}

// Validate returns an error if the session ID is not valid.
//
// The session ID is valid if the embedded peer ID is valid.
func (id SessionID) Validate() error {
	if id.Peer.Validate() == nil {
		return nil
	}

	return fmt.Errorf("session ID %s is invalid", id)
}

// At returns a SessionRef for this session ID.
func (id SessionID) At(rev RevisionNumber) SessionRef {
	return SessionRef{ID: id, Rev: rev}
}

// ShortString returns a string representation of the session ID based on the
// peer IDs short representation (e.g. "191C.45").
func (id SessionID) ShortString() string {
	return fmt.Sprintf("%s.%d", id.Peer.ShortString(), id.Seq)
}

// String returns a string representation of the session ID based on the full
// peer ID (e.g. "58AEE146-191C.45").
func (id SessionID) String() string {
	return fmt.Sprintf("%s.%d", id.Peer, id.Seq)
}

var sessionIDPattern *regexp.Regexp

func init() {
	sessionIDPattern = regexp.MustCompile(
		`^(.+)\-(.+)\.(.+)`,
	)
}
