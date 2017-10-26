package attributes

import "github.com/rinq/rinq-go/src/rinq"

// VList is a sequence of attributes with revision information.
type VList []VAttr

// Each calls fn for each attribute in the collection. Iteration stops
// when fn returns false.
func (l VList) Each(fn func(rinq.Attr) bool) {
	for _, attr := range l {
		if !fn(attr.Attr) {
			return
		}
	}
}

// IsEmpty returns true if there are no attributes in the collection.
func (l VList) IsEmpty() bool {
	return len(l) == 0
}

func (l VList) String() string {
	return ToString(l)
}
