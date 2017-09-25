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

	for _, attr := range t {
		if buf.Len() > 0 {
			buf.WriteString(", ")
		}

		buf.WriteString(attr.Attr.String())
	}

	return buf.String()
}
