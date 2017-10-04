package attrmeta

import "github.com/rinq/rinq-go/src/rinq/constraint"

// tableMatcher is a rinq.constraintVisitor that checks an constraint against an
// attribute table.
type tableMatcher struct {
	ns      string
	table   Table
	isMatch bool
}

func (m *tableMatcher) None() {
	m.isMatch = true
}

func (m *tableMatcher) Within(ns string, cons []constraint.Constraint) {
	for _, con := range cons {
		if !m.table.MatchConstraint(ns, con) {
			return
		}
	}

	m.isMatch = true
}

func (m *tableMatcher) Equal(k, v string) {
	a := m.table[m.ns][k]
	m.isMatch = a.Value == v
}

func (m *tableMatcher) NotEqual(k, v string) {
	a := m.table[m.ns][k]
	m.isMatch = a.Value != v
}

func (m *tableMatcher) Not(con constraint.Constraint) {
	m.isMatch = !m.table.MatchConstraint(m.ns, con)
}

func (m *tableMatcher) And(cons []constraint.Constraint) {
	for _, con := range cons {
		if !m.table.MatchConstraint(m.ns, con) {
			return
		}
	}

	m.isMatch = true
}

func (m *tableMatcher) Or(cons []constraint.Constraint) {
	for _, con := range cons {
		if m.table.MatchConstraint(m.ns, con) {
			m.isMatch = true
			return
		}
	}
}
