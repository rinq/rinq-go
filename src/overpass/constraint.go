package overpass

// Constraint represents a set of session attribute values used to determine
// which sessions receive multicast notifications.
type Constraint map[string]string

func (con Constraint) String() string {
	if len(con) == 0 {
		return "*"
	}

	str := ""
	for key, value := range con {
		if str != "" {
			str += ", "
		}

		if value == "" {
			str += "!" + key
		} else {
			str += key + "=" + value
		}
	}

	return str
}
