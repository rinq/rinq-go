package ident

type validatable interface {
	Validate() error
}

// MustValidate panics if v.Validate() returns an error.
func MustValidate(v validatable) {
	if err := v.Validate(); err != nil {
		panic(err)
	}
}
