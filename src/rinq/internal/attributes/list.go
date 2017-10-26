package attributes

import (
	"bytes"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/internal/x/bufferpool"
)

// List is a sequence of attributes.
type List []rinq.Attr

// WriteTo writes a representation of l to buf.
func (l List) WriteTo(buf *bytes.Buffer) {
	buf.WriteRune('{')

	for index, attr := range l {
		if index != 0 {
			buf.WriteString(", ")
		}

		buf.WriteString(attr.String())
	}

	buf.WriteRune('}')
}

func (l List) String() string {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	l.WriteTo(buf)

	return buf.String()
}
