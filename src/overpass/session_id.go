package overpass

import (
	"fmt"
	"regexp"
	"strconv"
)

// SessionID holds a unique session ID.
type SessionID struct {
	Peer PeerID
	Seq  uint32
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

// Validate returns nil if the ID is valid, it requires a non-zero seq.
func (id SessionID) Validate() error {
	if id.Peer.Validate() == nil {
		return nil
	}

	return fmt.Errorf("session ID %s is invalid", id)
}

// At creates a Ref from this ID.
func (id SessionID) At(rev RevisionNumber) SessionRef {
	return SessionRef{ID: id, Rev: rev}
}

// ShortString returns a string representation based on the peer's short string
// representation.
func (id SessionID) ShortString() string {
	return fmt.Sprintf("%s.%d", id.Peer.ShortString(), id.Seq)
}

func (id SessionID) String() string {
	return fmt.Sprintf("%s.%d", id.Peer, id.Seq)
}

var sessionIDPattern *regexp.Regexp

func init() {
	sessionIDPattern = regexp.MustCompile(
		`^(.+)\-(.+)\.(.+)`,
	)
}
