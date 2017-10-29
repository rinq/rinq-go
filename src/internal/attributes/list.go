package attributes

import (
	"github.com/rinq/rinq-go/src/rinq"
)

// List is a sequence of attributes.
type List []rinq.Attr

// Each calls fn for each attribute in the collection. Iteration stops
// when fn returns false.
func (l List) Each(fn func(rinq.Attr) bool) {
	for _, attr := range l {
		if !fn(attr) {
			return
		}
	}
}

// IsEmpty returns true if there are no attributes in the collection.
func (l List) IsEmpty() bool {
	return len(l) == 0
}

func (l List) String() string {
	return ToString(l)
}
