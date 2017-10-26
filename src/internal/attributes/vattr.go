package attributes

import (
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// VAttr is rinq.Attr with additional revision information.
type VAttr struct {
	rinq.Attr

	CreatedAt ident.Revision `json:"cr,omitempty"`
	UpdatedAt ident.Revision `json:"ur,omitempty"`
}
