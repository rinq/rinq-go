package attrmeta

import (
	"bytes"

	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/bufferpool"
)

// Diff is a sequence of attributes that have changed.
type Diff struct {
	Namespace string
	Revision  ident.Revision
	Attrs     List
}

// NewDiff returns a new Diff for the given namespace.
func NewDiff(ns string, rev ident.Revision, cap int) *Diff {
	return &Diff{
		Namespace: ns,
		Revision:  rev,
		Attrs:     make(List, 0, cap),
	}
}

// Append adds attributes to the diff.
func (d *Diff) Append(a ...Attr) {
	d.Attrs = append(d.Attrs, a...)
}

// IsEmpty returns true if the diff is empty.
func (d *Diff) IsEmpty() bool {
	return len(d.Attrs) == 0
}

// WriteTo writes a representation of d to buf.
func (d *Diff) WriteTo(buf *bytes.Buffer) {
	buf.WriteString(d.Namespace)
	buf.WriteString("::")
	d.WriteWithoutNamespaceTo(buf)
}

// WriteWithoutNamespaceTo writes a representation of d to buf, without the
// namespace name.
func (d *Diff) WriteWithoutNamespaceTo(buf *bytes.Buffer) {
	buf.WriteRune('{')

	for index, attr := range d.Attrs {
		if index != 0 {
			buf.WriteString(", ")
		}

		if attr.CreatedAt == d.Revision {
			buf.WriteString("+")
		}

		attr.WriteTo(buf)
	}

	buf.WriteRune('}')
}

func (d *Diff) String() string {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	d.WriteTo(buf)

	return buf.String()
}

// StringWithoutNamespace returns a string representation of d, without the
// namespace name.
func (d *Diff) StringWithoutNamespace() string {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	d.WriteWithoutNamespaceTo(buf)

	return buf.String()
}
