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

// Clone returns a copy of the attribute table.
func (t AttrTableWithMetaData) Clone() AttrTableWithMetaData {
	r := AttrTableWithMetaData{}

	for k, v := range t {
		r[k] = v
	}

	return r
}

// MatchConstraint returns true if the attributes match the given constraint.
func (t AttrTableWithMetaData) MatchConstraint(constraint overpass.Constraint) bool {
	for key, value := range constraint {
		if t[key].Value != value {
			return false
		}
	}

	return true
}
