package attributes

import "github.com/rinq/rinq-go/src/rinq"

// Table is a simple map-based implementation of rinq.AttrTable
type Table map[string]rinq.Attr

// Get returns the attribute with key k.
func (t Table) Get(k string) (rinq.Attr, bool) {
	attr, ok := t[k]
	return attr, ok
}

// Each calls fn for each attribute in the collection. Iteration stops
// when fn returns false.
func (t Table) Each(fn func(rinq.Attr) bool) {
	for _, attr := range t {
		if !fn(attr) {
			return
		}
	}
}

// IsEmpty returns true if there are no attributes in the table.
func (t Table) IsEmpty() bool {
	return len(t) == 0
}

// Len returns the number of attributes in the table.
func (t Table) Len() int {
	return len(t)
}
