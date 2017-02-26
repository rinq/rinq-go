package overpass

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

// AttrTable is a map of of attribute key to Attr.
type AttrTable map[string]Attr
