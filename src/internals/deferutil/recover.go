package deferutil

// Recover recovers from a panic, if the panic value is an error it is
// assigned to *err, otherwise it is propagated.
func Recover(err *error) {
	r := recover()

	if r == nil {
		return
	}

	if e, ok := r.(error); ok {
		*err = e
		return
	}

	panic(r)
}
