package attributes

import (
	"github.com/rinq/rinq-go/src/internal/x/bufferpool"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// Diff is an attribute collection representing a change to a set of attributes
// within a single namespace.
type Diff struct {
	VList

	Namespace string
	Revision  ident.Revision
}

// NewDiff returns a new Diff for the given namespace.
func NewDiff(ns string, rev ident.Revision) *Diff {
	return &Diff{
		Namespace: ns,
		Revision:  rev,
	}
}

// Append adds attributes to the diff.
func (d *Diff) Append(a ...VAttr) {
	d.VList = append(d.VList, a...)
}

func (d *Diff) String() string {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	buf.WriteString(d.Namespace)
	buf.WriteString("::")
	buf.WriteString(d.StringWithoutNamespace())

	return buf.String()
}

// StringWithoutNamespace returns a string representation of d, without the
// namespace name.
func (d *Diff) StringWithoutNamespace() string {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	buf.WriteRune('{')

	for index, attr := range d.VList {
		if index != 0 {
			buf.WriteString(", ")
		}

		if attr.CreatedAt == d.Revision {
			buf.WriteRune('+')
		}

		buf.WriteString(attr.String())
	}

	buf.WriteRune('}')

	return buf.String()
}
