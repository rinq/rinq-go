package attrmeta

import "github.com/over-pass/overpass-go/src/overpass"

// Table maps attribute keys to attributes with meta data.
type Table map[string]Attr

// Clone returns a copy of the attribute table.
func (t Table) Clone() Table {
	r := Table{}

	for k, v := range t {
		r[k] = v
	}

	return r
}

// MatchConstraint returns true if the attributes match the given constraint.
func (t Table) MatchConstraint(constraint overpass.Constraint) bool {
	for key, value := range constraint {
		if t[key].Value != value {
			return false
		}
	}

	return true
}
