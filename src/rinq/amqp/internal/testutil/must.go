package testutil

// Must panics if the right-most argument is a non-nil error.
func Must(v ...interface{}) {
	if len(v) == 0 {
		return
	}

	err, _ := v[len(v)-1].(error)
	if err != nil {
		panic(err)
	}
}
