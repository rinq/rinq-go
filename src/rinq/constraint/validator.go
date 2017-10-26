package constraint

import (
	"errors"

	"github.com/rinq/rinq-go/src/rinq/internal/namespaces"
)

type validator struct{}

func (v *validator) None(...interface{}) (interface{}, error) {
	return nil, nil
}

func (v *validator) Within(ns string, cons []Constraint, _ ...interface{}) (interface{}, error) {
	if err := namespaces.Validate(ns); err != nil {
		return nil, errors.New("WITHIN constraint has invalid namespace: " + err.Error())
	}

	if len(cons) == 0 {
		return nil, errors.New("WITHIN constraint has no terms")
	}

	for _, con := range cons {
		if _, err := con.Accept(v); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (v *validator) Equal(string, string, ...interface{}) (interface{}, error) {
	return nil, nil
}

func (v *validator) NotEqual(string, string, ...interface{}) (interface{}, error) {
	return nil, nil
}

func (v *validator) Not(con Constraint, _ ...interface{}) (interface{}, error) {
	return con.Accept(v)
}

func (v *validator) And(cons []Constraint, _ ...interface{}) (interface{}, error) {
	if len(cons) == 0 {
		return nil, errors.New("AND constraint has no terms")
	}

	for _, con := range cons {
		if _, err := con.Accept(v); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (v *validator) Or(cons []Constraint, _ ...interface{}) (interface{}, error) {
	if len(cons) == 0 {
		return nil, errors.New("OR constraint has no terms")
	}

	for _, con := range cons {
		if _, err := con.Accept(v); err != nil {
			return nil, err
		}
	}

	return nil, nil
}
