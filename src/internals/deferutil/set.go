package deferutil

// Set is a dynamic set of functions.
type Set struct {
	functions []func()
}

// Add a function to the set.
func (s *Set) Add(fn func()) {
	if fn != nil {
		s.functions = append(s.functions, fn)
	}
}

// Detach returns a function that invokes all functions in the set, then
// empties the set.
func (s *Set) Detach() func() {
	functions := s.functions
	s.functions = nil

	return func() {
		for _, fn := range functions {
			fn()
		}
	}
}

// Run invokes all functions in the set, unless Detach() has been called.
func (s *Set) Run() {
	for _, fn := range s.functions {
		fn()
	}
}
