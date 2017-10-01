package attrmeta

import (
	"bytes"

	"github.com/rinq/rinq-go/src/rinq/internal/bufferpool"
)

// List is a sequence of attributes with revision information.
type List []Attr

// WriteTo writes a representation of l to buf.
func (l List) WriteTo(buf *bytes.Buffer) {
	buf.WriteRune('{')

	for index, attr := range l {
		if index != 0 {
			buf.WriteString(", ")
		}

		attr.WriteTo(buf)
	}

	buf.WriteRune('}')
}

func (l List) String() string {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	l.WriteTo(buf)

	return buf.String()
}
