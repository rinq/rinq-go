package attributes

import (
	"github.com/rinq/rinq-go/src/rinq"
)

// VTable is a collection of attributes with revision information.
type VTable map[string]VAttr

// Each calls fn for each attribute in the collection. Iteration stops
// when fn returns false.
func (t VTable) Each(fn func(rinq.Attr) bool) {
	for _, attr := range t {
		if !fn(attr.Attr) {
			return
		}
	}
}

// IsEmpty returns true if there are no attributes in the table.
func (t VTable) IsEmpty() bool {
	return len(t) == 0
}

func (t VTable) String() string {
	return ToString(t)
}

// Clone returns a copy of t.
func (t VTable) Clone() VTable {
	c := VTable{}

	for k, v := range t {
		c[k] = v
	}

	return c
}
