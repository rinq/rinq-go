package constraint

type op string

const (
	noneOp     op = "*"
	withinOp   op = "ns"
	equalOp    op = "="
	notEqualOp op = "!="
	notOp      op = "!"
	andOp      op = "&"
	orOp       op = "|"
)

// Visitor is used to walk a constraint hierarchy.
type Visitor interface {
	None(args ...interface{}) (interface{}, error)
	Within(ns string, cons []Constraint, args ...interface{}) (interface{}, error)
	Equal(k, v string, args ...interface{}) (interface{}, error)
	NotEqual(k, v string, args ...interface{}) (interface{}, error)
	Not(con Constraint, args ...interface{}) (interface{}, error)
	And(cons []Constraint, args ...interface{}) (interface{}, error)
	Or(cons []Constraint, args ...interface{}) (interface{}, error)
}
