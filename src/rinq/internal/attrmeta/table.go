package attrmeta

import (
	"bytes"

	"github.com/rinq/rinq-go/src/rinq/internal/bufferpool"
)

// Table maps namespace to attribute table.
type Table map[string]Namespace

// CloneAndMerge returns a copy of the attribute table with the n namespaced replaced
// with ns. ns is NOT cloned.
func (t Table) CloneAndMerge(name string, ns Namespace) Table {
	r := Table{name: ns}

	for n, a := range t {
		if n != name {
			r[n] = a.Clone()
		}
	}

	return r
}

// WriteTo writes a respresentation of t to buf.
// Non-frozen attributes with empty-values are omitted.
func (t Table) WriteTo(buf *bytes.Buffer) {
	sub := bufferpool.Get()
	defer bufferpool.Put(sub)

	first := true
	for n, a := range t {
		sub.Reset()

		if a.WriteWithNameTo(sub, n) {
			if first {
				first = false
			} else {
				buf.WriteRune(' ')
			}

			_, _ = sub.WriteTo(buf)
		}
	}

	if first {
		buf.WriteString("{}")
	}
}

func (t Table) String() string {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	t.WriteTo(buf)

	return buf.String()
}
