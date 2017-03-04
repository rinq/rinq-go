package attrmeta

import (
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// Attr is rinq.Attr with additional meta data.
type Attr struct {
	rinq.Attr

	CreatedAt ident.Revision `json:"cr,omitempty"`
	UpdatedAt ident.Revision `json:"ur,omitempty"`
}
