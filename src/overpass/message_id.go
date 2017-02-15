package overpass

import (
	"fmt"
	"regexp"
	"strconv"
)

// MessageID uniquely identifies a message that originated from a session.
type MessageID struct {
	Session SessionRef
	Seq     uint32
}

// ParseMessageID parses a string representation of a message ID.
func ParseMessageID(str string) (id MessageID, err error) {
	matches := messageIDPattern.FindStringSubmatch(str)

	if len(matches) != 0 {
		// Read the peer ID clock component ...
		var value uint64
		value, err = strconv.ParseUint(matches[1], 16, 64)
		if err != nil {
			return
		}
		id.Session.ID.Peer.Clock = value

		// Read the peer ID random component ...
		value, err = strconv.ParseUint(matches[2], 16, 16)
		if err != nil {
			return
		}
		id.Session.ID.Peer.Rand = uint16(value)

		// Read the session ID sequence component ...
		value, err = strconv.ParseUint(matches[3], 10, 32)
		if err != nil {
			return
		}
		id.Session.ID.Seq = uint32(value)

		// Read the session version ...
		value, err = strconv.ParseUint(matches[4], 10, 32)
		if err != nil {
			return
		}
		id.Session.Rev = RevisionNumber(value)

		// Read the message ID sequence component ...
		value, err = strconv.ParseUint(matches[5], 10, 32)
		if err != nil {
			return
		}
		id.Seq = uint32(value)
	}

	err = id.Validate()
	return
}

// Validate returns nil if the ID is valid.
func (id MessageID) Validate() error {
	if id.Session.Validate() == nil && id.Seq != 0 {
		return nil
	}

	return fmt.Errorf("Message ID %s is invalid.", id)
}

// ShortString returns a string representation based on the session's short
// string representation.
func (id MessageID) ShortString() string {
	return fmt.Sprintf("%s#%d", id.Session.ShortString(), id.Seq)
}

func (id MessageID) String() string {
	return fmt.Sprintf("%s#%d", id.Session, id.Seq)
}

var messageIDPattern *regexp.Regexp

func init() {
	messageIDPattern = regexp.MustCompile(
		`^(.+)\-(.+)\.(.+)@(.+)#(.+)$`,
	)
}
