package attrmeta

import "github.com/rinq/rinq-go/src/rinq"

// matcher is a rinq.constraintVisitor that checks an constraint against an
// attribute table.
type matcher struct {
	ns      string
	table   Table
	isMatch bool
}

func (m *matcher) Within(ns string, cons []rinq.Constraint) {
	for _, con := range cons {
		if !m.table.MatchConstraint(ns, con) {
			return
		}
	}

	m.isMatch = true
}

func (m *matcher) Equal(k string, vals []string) {
	a := m.table[m.ns][k]

	for _, v := range vals {
		if a.Value == v {
			m.isMatch = true
			return
		}
	}
}

func (m *matcher) NotEqual(k string, v []string) {
	a := m.table[m.ns][k]

	for _, x := range v {
		if a.Value == x {
			return
		}
	}

	m.isMatch = true
}

func (m *matcher) Empty(k string) {
	a := m.table[m.ns][k]
	m.isMatch = a.Value == ""
}

func (m *matcher) NotEmpty(k string) {
	a := m.table[m.ns][k]
	m.isMatch = a.Value != ""
}

func (m *matcher) Not(con rinq.Constraint) {
	m.isMatch = !m.table.MatchConstraint(m.ns, con)
}

func (m *matcher) And(cons []rinq.Constraint) {
	for _, con := range cons {
		if !m.table.MatchConstraint(m.ns, con) {
			return
		}
	}

	m.isMatch = true
}

func (m *matcher) Or(cons []rinq.Constraint) {
	for _, con := range cons {
		if m.table.MatchConstraint(m.ns, con) {
			m.isMatch = true
			return
		}
	}
}
