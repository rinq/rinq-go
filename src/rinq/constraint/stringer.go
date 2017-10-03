package constraint

import (
	"bytes"

	"github.com/rinq/rinq-go/src/rinq/internal/strutil"
)

type stringer struct {
	buf       *bytes.Buffer
	hasBraces []bool
}

func (s *stringer) Within(ns string, cons []Constraint) {
	s.buf.WriteString(ns)
	s.buf.WriteString("::")
	s.join(", ", cons)
}

func (s *stringer) Equal(k, v string) {
	s.open()

	if v == "" {
		s.buf.WriteRune('!')
		s.buf.WriteString(strutil.Escape(k))
	} else {
		s.buf.WriteString(strutil.Escape(k))
		s.buf.WriteRune('=')
		s.buf.WriteString(strutil.Escape(v))
	}

	s.close()
}

func (s *stringer) NotEqual(k, v string) {
	s.open()

	if v == "" {
		s.buf.WriteString(strutil.Escape(k))
	} else {
		s.buf.WriteString(strutil.Escape(k))
		s.buf.WriteString("!=")
		s.buf.WriteString(strutil.Escape(v))
	}

	s.close()
}

func (s *stringer) Not(con Constraint) {
	s.open()
	s.buf.WriteString("! ")
	con.Accept(s)
	s.close()
}

func (s *stringer) And(cons []Constraint) {
	s.join(", ", cons)
}

func (s *stringer) Or(cons []Constraint) {
	s.join("|", cons)
}

func (s *stringer) join(sep string, cons []Constraint) {
	if len(cons) == 1 {
		cons[0].Accept(s)
		return
	}

	s.buf.WriteRune('{')
	s.push(true)

	for i, con := range cons {
		if i != 0 {
			s.buf.WriteString(sep)
		}
		con.Accept(s)
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
