package rinq

import (
	"bytes"

	"github.com/rinq/rinq-go/src/rinq/internal/strutil"
)

type constraintStringer struct {
	buf *bytes.Buffer
}

func (s *constraintStringer) Within(ns string, cons []Constraint) {
	s.buf.WriteString(ns)
	s.buf.WriteString("::{")
	s.join(", ", cons)
	s.buf.WriteString("}")
}

func (s *constraintStringer) Equal(k string, vals []string) {
	if len(vals) == 1 {
		s.buf.WriteString(strutil.Escape(k))
		s.buf.WriteRune('=')
		s.buf.WriteString(strutil.Escape(vals[0]))
		return
	}

	s.buf.WriteString(strutil.Escape(k))
	s.buf.WriteString(" IN (")
	for i, v := range vals {
		if i != 0 {
			s.buf.WriteString(", ")
		}
		s.buf.WriteString(strutil.Escape(v))
	}
	s.buf.WriteRune(')')
}

func (s *constraintStringer) NotEqual(k string, vals []string) {
	if len(vals) == 1 {
		s.buf.WriteString(strutil.Escape(k))
		s.buf.WriteString("!=")
		s.buf.WriteString(strutil.Escape(vals[0]))
		return
	}

	s.buf.WriteString(strutil.Escape(k))
	s.buf.WriteString(" NOT IN (")
	for i, v := range vals {
		if i != 0 {
			s.buf.WriteString(", ")
		}
		s.buf.WriteString(strutil.Escape(v))
	}
	s.buf.WriteRune(')')
}

func (s *constraintStringer) Empty(k string) {
	s.buf.WriteRune('!')
	s.buf.WriteString(strutil.Escape(k))
}

func (s *constraintStringer) NotEmpty(k string) {
	s.buf.WriteString(strutil.Escape(k))
}

func (s *constraintStringer) Not(con Constraint) {
	s.buf.WriteString("NOT ")
	con.Accept(s)
}

func (s *constraintStringer) And(cons []Constraint) {
	if len(cons) == 1 {
		cons[0].Accept(s)
	} else {
		s.buf.WriteRune('{')
		s.join(", ", cons)
		s.buf.WriteRune('}')
	}
}

func (s *constraintStringer) Or(cons []Constraint) {
	if len(cons) == 1 {
		cons[0].Accept(s)
	} else {
		s.buf.WriteRune('{')
		s.join("|", cons)
		s.buf.WriteRune('}')
	}
}

func (s *constraintStringer) join(sep string, cons []Constraint) {
	for i, con := range cons {
		if i != 0 {
			s.buf.WriteString(sep)
		}
		con.Accept(s)
	}
}
