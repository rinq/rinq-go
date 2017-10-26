package attributes

import "github.com/rinq/rinq-go/src/rinq"

// Iterator is an interface for enumerating attributes.
type Iterator interface {
	Each(func(rinq.Attr) bool)
}

// ToMap returns a new map of attributes from i
func ToMap(i Iterator) map[string]rinq.Attr {
	m := map[string]rinq.Attr{}

	i.Each(func(a rinq.Attr) bool {
		m[a.Key] = a
		return true
	})

	return m
}
