package attrmeta

import (
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/internal/bufferpool"
)

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
func (t Table) MatchConstraint(constraint rinq.Constraint) bool {
	for key, value := range constraint {
		if t[key].Value != value {
			return false
		}
	}

	return true
}

func (t Table) String() string {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	WriteTable(buf, t)

	return buf.String()
}

// NamespacedTable maps namespace to attribute table.
type NamespacedTable map[string]Table

// CloneAndReplace returns a copy of the attribute table with the ns namespaced replaced
// with nt.
func (t NamespacedTable) CloneAndReplace(ns string, nt Table) NamespacedTable {
	r := NamespacedTable{ns: nt}

	for n, a := range t {
		if n != ns {
			r[n] = a.Clone()
		}
	}

	return r
}

func (t NamespacedTable) String() string {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	WriteNamespacedTable(buf, t)

	return buf.String()
}
