package rinq

import "github.com/rinq/rinq-go/src/rinq/internal/bufferpool"

// Constraint represents a set of session attribute values used to determine
// which sessions receive multicast notifications.
type Constraint map[string]string

func (con Constraint) String() string {
	if len(con) == 0 {
		return "*"
	}

	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	for key, value := range con {
		if buf.Len() > 0 {
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

	return buf.String()
}
