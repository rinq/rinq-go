package rinq

import "github.com/rinq/rinq-go/src/rinq/internal/x/bufferpool"

// Attr is a sesssion attribute.
//
// Sessions contain a versioned key/value store. See the Session interface for
// more information.
type Attr struct {
	// Key is an application-defined identifier for the attribute. Keys are
	// unique within a session. Any valid UTF-8 string can be used a key,
	// including the empty string.
	Key string `json:"k"`

	// Value is the attribute's value. Any valid UTF-8 string can be used as a
	// value, including the empty string.
	Value string `json:"v,omitempty"`

	// IsFrozen is true if the attribute is "frozen" such that it can never be
	// altered again (for a given session).
	IsFrozen bool `json:"f,omitempty"`
}

// Set is a convenience method that creates an Attr with the specified key and
// value.
func Set(key, value string) Attr {
	return Attr{Key: key, Value: value}
}

// Freeze is a convenience method that returns an Attr with the specified key
// and value, and the IsFrozen flag set to true.
func Freeze(key, value string) Attr {
	return Attr{Key: key, Value: value, IsFrozen: true}
}

func (attr Attr) String() string {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	if attr.Value == "" {
		if attr.IsFrozen {
			buf.WriteString("!")
		} else {
			buf.WriteString("-")
		}
		buf.WriteString(attr.Key)
	} else {
		buf.WriteString(attr.Key)
		if attr.IsFrozen {
			buf.WriteString("@")
		} else {
			buf.WriteString("=")
		}
		buf.WriteString(attr.Value)
	}

	return buf.String()
}

// AttrTable is a read-only table of session attributes.
type AttrTable interface {
	// Get returns the attribute with key k.
	Get(k string) (Attr, bool)

	// Each calls fn for each attribute in the collection. Iteration stops
	// when fn returns false.
	Each(fn func(Attr) bool)

	// IsEmpty returns true if there are no attributes in the table.
	IsEmpty() bool

	String() string
}
