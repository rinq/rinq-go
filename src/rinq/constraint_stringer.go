package rinq

import (
	"bytes"

	"github.com/rinq/rinq-go/src/rinq/internal/strutil"
)

type constraintStringer struct {
	buf       *bytes.Buffer
	hasBraces []bool
}

func (s *constraintStringer) Within(ns string, cons []Constraint) {
	s.buf.WriteString(ns)
	s.buf.WriteString("::")
	s.join(", ", cons)
}

func (s *constraintStringer) Equal(k, v string) {
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

func (s *constraintStringer) NotEqual(k, v string) {
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

func (s *constraintStringer) Not(con Constraint) {
	s.open()
	s.buf.WriteString("! ")
	con.Accept(s)
	s.close()
}

func (s *constraintStringer) And(cons []Constraint) {
	s.join(", ", cons)
}

func (s *constraintStringer) Or(cons []Constraint) {
	s.join("|", cons)
}

func (s *constraintStringer) join(sep string, cons []Constraint) {
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

func (s *constraintStringer) push(b bool) {
	s.hasBraces = append(s.hasBraces, b)
}

func (s *constraintStringer) pop() {
	s.hasBraces = s.hasBraces[:len(s.hasBraces)-1]
}

func (s *constraintStringer) needsBraces() bool {
	if len(s.hasBraces) == 0 {
		return true
	}

	return !s.hasBraces[len(s.hasBraces)-1]
}

func (s *constraintStringer) open() {
	if s.needsBraces() {
		s.buf.WriteRune('{')
	}

	s.push(true)
}

func (s *constraintStringer) close() {
	s.pop()

	if s.needsBraces() {
		s.buf.WriteRune('}')
	}
}
