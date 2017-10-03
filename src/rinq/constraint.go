package rinq

import "github.com/rinq/rinq-go/src/rinq/internal/bufferpool"

// Constraint is a boolean expression evaluaed against session attribute values
// to determine which sessions receive a multicast notification.
//
// See Session.NotifyMany() to send a multicast notification.
type Constraint struct {
	Op       constraintOp `json:"o,omitempty"`
	Children []Constraint `json:"c,omitempty"`
	Key      string       `json:"k,omitempty"`
	Values   []string     `json:"v,omitempty"`
}

func (c Constraint) String() string {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	c.Accept(&constraintStringer{buf})

	return buf.String()
}

// And returns a Constraint that evaluates to true if both c and con evaluate to
// true.
func (c Constraint) And(con Constraint) Constraint {
	return And(c, con)
}

// Or returns a Constraint that evaluates to true if at least one of c and con
// evaluate to true.
func (c Constraint) Or(con Constraint) Constraint {
	return Or(c, con)
}

// Within returns a Constraint that evaluates to true when each constraint in
// cons evaluates to true within the ns namespace.
func Within(ns string, cons ...Constraint) Constraint {
	return Constraint{
		Op:       withinOp,
		Children: cons,
		Key:      ns,
	}
}

// Equal returns a Constraint that evaluates to true when the attribute k is equal
// to any of the values in vals.
func Equal(k string, v ...string) Constraint {
	return Constraint{
		Op:     equalOp,
		Key:    k,
		Values: v,
	}
}

// NotEqual returns a Constraint that evaluates to true when the attribute k is not
// equal to any of the values in vals.
func NotEqual(k string, v ...string) Constraint {
	return Constraint{
		Op:     notEqualOp,
		Key:    k,
		Values: v,
	}
}

// Empty returns a Constraint that evaluates to true when the attribute k has a
// value equal to the empty string.
func Empty(k string) Constraint {
	return Constraint{
		Op:  emptyOp,
		Key: k,
	}
}

// NotEmpty returns a Constraint that evaluates to true when the attribute k has
// a value not equal to the empty string.
func NotEmpty(k string) Constraint {
	return Constraint{
		Op:  notEmptyOp,
		Key: k,
	}
}

// Not returns a Constraint  that evaluates to true when e evaluates to false,
// and vice-versa.
func Not(con Constraint) Constraint {
	return Constraint{
		Op:       notOp,
		Children: []Constraint{con},
	}
}

// And returns a Constraint that evaluates to true when all constraints in cons
// evaluate to true.
func And(cons ...Constraint) Constraint {
	return Constraint{
		Op:       andOp,
		Children: cons,
	}
}

// Or returns a Constraint that evaluates to true when one or more of the
// constraints in cons evaluate to true.
func Or(cons ...Constraint) Constraint {
	return Constraint{
		Op:       orOp,
		Children: cons,
	}
}

// Accept calls the method on v that corresponds to the operation type of c.
func (c Constraint) Accept(v ConstraintVisitor) {
	switch c.Op {
	case withinOp:
		v.Within(c.Key, c.Children)
	case equalOp:
		v.Equal(c.Key, c.Values)
	case notEqualOp:
		v.NotEqual(c.Key, c.Values)
	case emptyOp:
		v.Empty(c.Key)
	case notEmptyOp:
		v.NotEmpty(c.Key)
	case notOp:
		v.Not(c.Children[0])
	case andOp:
		v.And(c.Children)
	case orOp:
		v.Or(c.Children)
	default:
		panic("unrecognized constraint operation: " + c.Op)
	}
}

type constraintOp string

const (
	withinOp   constraintOp = "ns"
	equalOp    constraintOp = "="
	notEqualOp constraintOp = "!="
	emptyOp    constraintOp = "-"
	notEmptyOp constraintOp = "+"
	notOp      constraintOp = "!"
	andOp      constraintOp = "&"
	orOp       constraintOp = "|"
)

// ConstraintVisitor is used to walk a constraint hierarchy.
type ConstraintVisitor interface {
	Within(ns string, cons []Constraint)
	Equal(k string, v []string)
	NotEqual(k string, v []string)
	Empty(k string)
	NotEmpty(k string)
	Not(con Constraint)
	And(cons []Constraint)
	Or(cons []Constraint)
}
