package overpass

// Set returns a new attribute.
func Set(key, value string) Attr {
	return Attr{Key: key, Value: value}
}

// Freeze returns a new frozen attribute.
func Freeze(key, value string) Attr {
	return Attr{Key: key, Value: value, IsFrozen: true}
}

// Attr holds attribute content and meta-data.
// Keys and values are UTF-8 strings.
type Attr struct {
	Key      string
	Value    string
	IsFrozen bool
}

// AttrTable maps attribute keys to attributes.
type AttrTable map[string]Attr
