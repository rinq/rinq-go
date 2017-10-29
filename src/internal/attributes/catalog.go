package attributes

import (
	"github.com/rinq/rinq-go/src/internal/x/bufferpool"
	"github.com/rinq/rinq-go/src/rinq/constraint"
)

// Catalog is a namespaced collection of attributes.
type Catalog map[string]VTable

// WithNamespace returns a copy of the catalog with the ns namespace replaced by t.
func (c Catalog) WithNamespace(ns string, t VTable) Catalog {
	r := Catalog{ns: t}

	for n, t := range c {
		if n != ns {
			r[n] = t.Clone()
		}
	}

	return r
}

// MatchConstraint returns true if con evalutes to true for the attributes in
// attrs. The ns namespace is the default namespace used if there is no 'within'
// constraint.
func (c Catalog) MatchConstraint(ns string, con constraint.Constraint) bool {
	isMatch, _ := con.Accept(&catalogMatcher{c}, ns)
	return isMatch.(bool)
}

// IsEmpty returns true if there are no attributes in the catalog.
func (c Catalog) IsEmpty() bool {
	for _, t := range c {
		if !t.IsEmpty() {
			return false
		}
	}

	return true
}

func (c Catalog) String() string {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	empty := true
	for ns, t := range c {
		if t.IsEmpty() {
			continue
		}

		if empty {
			empty = false
		} else {
			buf.WriteRune(' ')
		}

		buf.WriteString(ns)
		buf.WriteString("::")
		buf.WriteString(t.String())
	}

	if empty {
		return "{}"
	}

	return buf.String()
}
