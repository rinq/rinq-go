package rinq

import "github.com/rinq/rinq-go/src/rinq/internal/bufferpool"

// Constraint represents a set of session attribute values used to determine
// which sessions receive a multicast notification.
//
// See Session.NotifyMany() to send a multicast notification.
type Constraint map[string]string

func (con Constraint) String() string {
	if len(con) == 0 {
		return "{*}"
	}

	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	buf.WriteRune('{')

	first := true
	for key, value := range con {
		if first {
			first = false
		} else {
			buf.WriteString(", ")
		}

		if value == "" {
			buf.WriteRune('!')
			buf.WriteString(key)
		} else {
			buf.WriteString(key)
			buf.WriteRune('=')
			buf.WriteString(value)
		}
	}

	buf.WriteRune('}')

	return buf.String()
}
