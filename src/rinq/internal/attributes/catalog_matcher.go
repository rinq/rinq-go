package attributes

import "github.com/rinq/rinq-go/src/rinq/constraint"

// catalogMatcher is a constraint.Visitor that evaluates a constraint against an
// attribute catalog.
type catalogMatcher struct {
	cat Catalog
}

// unpackNamespace extracts the first element of args as a string.
func unpackNamespace(args []interface{}) string {
	return args[0].(string)
}

func (m *catalogMatcher) None(_ ...interface{}) (interface{}, error) {
	return true, nil
}

func (m *catalogMatcher) Within(ns string, cons []constraint.Constraint, _ ...interface{}) (interface{}, error) {
	for _, con := range cons {
		isMatch, _ := con.Accept(m, ns)
		if !isMatch.(bool) {
			return false, nil
		}
	}

	return true, nil
}

func (m *catalogMatcher) Equal(k, v string, args ...interface{}) (interface{}, error) {
	ns := unpackNamespace(args)
	return m.cat[ns][k].Value == v, nil
}

func (m *catalogMatcher) NotEqual(k, v string, args ...interface{}) (interface{}, error) {
	ns := unpackNamespace(args)
	return m.cat[ns][k].Value != v, nil
}

func (m *catalogMatcher) Not(con constraint.Constraint, args ...interface{}) (interface{}, error) {
	isMatch, _ := con.Accept(m, args...)
	return !isMatch.(bool), nil
}

func (m *catalogMatcher) And(cons []constraint.Constraint, args ...interface{}) (interface{}, error) {
	for _, con := range cons {
		isMatch, _ := con.Accept(m, args...)
		if !isMatch.(bool) {
			return false, nil
		}
	}

	return true, nil
}

func (m *catalogMatcher) Or(cons []constraint.Constraint, args ...interface{}) (interface{}, error) {
	for _, con := range cons {
		isMatch, _ := con.Accept(m, args...)
		if isMatch.(bool) {
			return true, nil
		}
	}

	return false, nil
}
