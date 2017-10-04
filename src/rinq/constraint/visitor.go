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
	None()
	Within(ns string, cons []Constraint)
	Equal(k, v string)
	NotEqual(k, v string)
	Not(con Constraint)
	And(cons []Constraint)
	Or(cons []Constraint)
}
