package constraint

import "github.com/rinq/rinq-go/src/rinq/internal/bufferpool"

// Constraint is a boolean expression evaluated against session attribute values
// to determine which sessions receive a multicast notification.
//
// See Session.NotifyMany() to send a multicast notification.
type Constraint struct {
	Op    op           `json:"o,omitempty"`
	Terms []Constraint `json:"t,omitempty"`
	Key   string       `json:"k,omitempty"`
	Value string       `json:"v,omitempty"`
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

// Validate returns nil if c is a valid constraint.
func (c Constraint) Validate() (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				panic(r)
			}
		}
	}()

	v := &validator{}
	c.Accept(v)

	return
}

// Accept calls the method on v that corresponds to the operation type of c.
func (c Constraint) Accept(v Visitor) {
	switch c.Op {
	case noneOp:
		v.None()
	case withinOp:
		v.Within(c.Value, c.Terms)
	case equalOp:
		v.Equal(c.Key, c.Value)
	case notEqualOp:
		v.NotEqual(c.Key, c.Value)
	case notOp:
		v.Not(c.Terms[0])
	case andOp:
		v.And(c.Terms)
	case orOp:
		v.Or(c.Terms)
	default:
		panic("unrecognized constraint operation: " + c.Op)
	}
}

func (c Constraint) String() string {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	v := &stringer{buf, nil}
	c.Accept(v)

	return buf.String()
}

// None is a Constraint that always evaluates to true, and hence provides
// "no constraint" on the sessions that receive the notification.
var None = Constraint{Op: noneOp}

// Within returns a Constraint that evaluates to true when each constraint in
// cons evaluates to true within the ns namespace.
func Within(ns string, cons ...Constraint) Constraint {
	return Constraint{
		Op:    withinOp,
		Terms: cons,
		Value: ns,
	}
}

// Equal returns a Constraint that evaluates to true when the attribute k is
// equal to v.
func Equal(k, v string) Constraint {
	return Constraint{
		Op:    equalOp,
		Key:   k,
		Value: v,
	}
}

// NotEqual returns a Constraint that evaluates to true when the attribute k is
// not equal to v.
func NotEqual(k, v string) Constraint {
	return Constraint{
		Op:    notEqualOp,
		Key:   k,
		Value: v,
	}
}

// Empty returns a Constraint that evaluates to true when the attribute k has a
// value equal to the empty string.
func Empty(k string) Constraint {
	return Constraint{
		Op:  equalOp,
		Key: k,
	}
}

// NotEmpty returns a Constraint that evaluates to true when the attribute k has
// a value not equal to the empty string.
func NotEmpty(k string) Constraint {
	return Constraint{
		Op:  notEqualOp,
		Key: k,
	}
}

// Not returns a Constraint that evaluates to true when e evaluates to false,
// and vice-versa.
func Not(con Constraint) Constraint {
	return Constraint{
		Op:    notOp,
		Terms: []Constraint{con},
	}
}

// And returns a Constraint that evaluates to true when all constraints in cons
// evaluate to true.
func And(cons ...Constraint) Constraint {
	return Constraint{
		Op:    andOp,
		Terms: cons,
	}
}

// Or returns a Constraint that evaluates to true when one or more of the
// constraints in cons evaluate to true.
func Or(cons ...Constraint) Constraint {
	return Constraint{
		Op:    orOp,
		Terms: cons,
	}
}
