package attrmeta

import "github.com/over-pass/overpass-go/src/overpass"

// Attr is overpass.Attr with additional meta data.
type Attr struct {
	overpass.Attr

	CreatedAt overpass.RevisionNumber `json:"cr,omitempty"`
	UpdatedAt overpass.RevisionNumber `json:"ur,omitempty"`
}
