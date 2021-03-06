package constraint

import (
	"bytes"

	"github.com/rinq/rinq-go/src/internal/x/repr"
)

type stringer struct {
	buf       *bytes.Buffer
	hasBraces []bool
}

func (s *stringer) None(...interface{}) (interface{}, error) {
	s.open()
	s.buf.WriteString("*")
	s.close()

	return nil, nil
}

func (s *stringer) Within(ns string, cons []Constraint, _ ...interface{}) (interface{}, error) {
	s.buf.WriteString(ns)
	s.buf.WriteString("::")
	s.join(", ", cons)

	return nil, nil
}

func (s *stringer) Equal(k, v string, _ ...interface{}) (interface{}, error) {
	s.open()

	if v == "" {
		s.buf.WriteRune('!')
		s.buf.WriteString(repr.Escape(k))
	} else {
		s.buf.WriteString(repr.Escape(k))
		s.buf.WriteRune('=')
		s.buf.WriteString(repr.Escape(v))
	}

	s.close()

	return nil, nil
}

func (s *stringer) NotEqual(k, v string, _ ...interface{}) (interface{}, error) {
	s.open()

	if v == "" {
		s.buf.WriteString(repr.Escape(k))
	} else {
		s.buf.WriteString(repr.Escape(k))
		s.buf.WriteString("!=")
		s.buf.WriteString(repr.Escape(v))
	}

	s.close()

	return nil, nil
}

func (s *stringer) Not(con Constraint, _ ...interface{}) (interface{}, error) {
	s.open()
	s.buf.WriteString("! ")
	_, _ = con.Accept(s)
	s.close()

	return nil, nil
}

func (s *stringer) And(cons []Constraint, _ ...interface{}) (interface{}, error) {
	s.join(", ", cons)

	return nil, nil
}

func (s *stringer) Or(cons []Constraint, _ ...interface{}) (interface{}, error) {
	s.join("|", cons)

	return nil, nil
}

func (s *stringer) join(sep string, cons []Constraint) {
	if len(cons) == 1 {
		_, _ = cons[0].Accept(s)
		return
	}

	s.buf.WriteRune('{')
	s.push(true)

	for i, con := range cons {
		if i != 0 {
			s.buf.WriteString(sep)
		}
		_, _ = con.Accept(s)
	}

	s.pop()
	s.buf.WriteRune('}')
}

func (s *stringer) push(b bool) {
	s.hasBraces = append(s.hasBraces, b)
}

func (s *stringer) pop() {
	s.hasBraces = s.hasBraces[:len(s.hasBraces)-1]
}

func (s *stringer) needsBraces() bool {
	if len(s.hasBraces) == 0 {
		return true
	}

	return !s.hasBraces[len(s.hasBraces)-1]
}

func (s *stringer) open() {
	if s.needsBraces() {
		s.buf.WriteRune('{')
	}

	s.push(true)
}

func (s *stringer) close() {
	s.pop()

	if s.needsBraces() {
		s.buf.WriteRune('}')
	}
}
