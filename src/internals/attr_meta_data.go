package internals

import "github.com/over-pass/overpass-go/src/overpass"

// AttrWithMetaData is an attribute with additional meta data.
type AttrWithMetaData struct {
	overpass.Attr
	CreatedAt overpass.RevisionNumber
	UpdatedAt overpass.RevisionNumber
}

// AttrTableWithMetaData maps attribute keys to attributes with meta data.
type AttrTableWithMetaData map[string]AttrWithMetaData
