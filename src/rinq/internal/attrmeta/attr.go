package attrmeta

import (
	"bytes"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// Attr is rinq.Attr with additional revision information.
type Attr struct {
	rinq.Attr

	CreatedAt ident.Revision `json:"cr,omitempty"`
	UpdatedAt ident.Revision `json:"ur,omitempty"`
}

// WriteTo writes a representation of a to the buf.
func (a *Attr) WriteTo(buf *bytes.Buffer) {
	if a.Value == "" {
		if a.IsFrozen {
			buf.WriteString("!")
		} else {
			buf.WriteString("-")
		}
		buf.WriteString(a.Key)
	} else {
		buf.WriteString(a.Key)
		if a.IsFrozen {
			buf.WriteString("@")
		} else {
			buf.WriteString("=")
		}
		buf.WriteString(a.Value)
	}
}
