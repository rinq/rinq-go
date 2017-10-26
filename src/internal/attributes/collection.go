package attributes

import (
	"github.com/rinq/rinq-go/src/internal/x/bufferpool"
	"github.com/rinq/rinq-go/src/rinq"
)

// Collection is a collection of attributes.
type Collection interface {
	// Each calls fn for each attribute in the collection. Iteration stops
	// when fn returns false.
	Each(fn func(rinq.Attr) bool)

	// IsEmpty returns true if there are no attributes in the collection.
	IsEmpty() bool

	String() string
}

// ToMap returns a new map of attributes from attrs.
func ToMap(attrs Collection) map[string]rinq.Attr {
	m := map[string]rinq.Attr{}

	attrs.Each(func(a rinq.Attr) bool {
		m[a.Key] = a
		return true
	})

	return m
}

// ToString provides an implementation of Collection.String() using
// Collection.Each().
func ToString(attrs Collection) string {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	buf.WriteRune('{')

	empty := true
	attrs.Each(func(attr rinq.Attr) bool {
		if empty {
			empty = false
		} else {
			buf.WriteString(", ")
		}

		buf.WriteString(attr.String())
		return true
	})

	buf.WriteRune('}')

	return buf.String()
}
