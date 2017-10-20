package constraint

import (
	"errors"

	"github.com/rinq/rinq-go/src/rinq/internal/namespaces"
)

type validator struct{}

func (v *validator) None() {
}

func (v *validator) Within(ns string, cons []Constraint) {
	if err := namespaces.Validate(ns); err != nil {
		panic(errors.New("WITHIN constraint has invalid namespace: " + err.Error()))
	}

	if len(cons) == 0 {
		panic(errors.New("WITHIN constraint has no terms"))
	}

	for _, con := range cons {
		con.Accept(v)
	}
}

func (v *validator) Equal(string, string) {
}

func (v *validator) NotEqual(string, string) {
}

func (v *validator) Not(con Constraint) {
	con.Accept(v)
}

func (v *validator) And(cons []Constraint) {
	if len(cons) == 0 {
		panic(errors.New("AND constraint has no terms"))
	}

	for _, con := range cons {
		con.Accept(v)
	}
}

func (v *validator) Or(cons []Constraint) {
	if len(cons) == 0 {
		panic(errors.New("OR constraint has no terms"))
	}

	for _, con := range cons {
		con.Accept(v)
	}
}
