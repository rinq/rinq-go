package attrmeta

import "github.com/rinq/rinq-go/src/rinq"

// Attr is rinq.Attr with additional meta data.
type Attr struct {
	rinq.Attr

	CreatedAt rinq.RevisionNumber `json:"cr,omitempty"`
	UpdatedAt rinq.RevisionNumber `json:"ur,omitempty"`
}
