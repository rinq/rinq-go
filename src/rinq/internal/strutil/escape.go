package strutil

import (
	"encoding/json"
	"strings"
)

// Escape returns human-readable, possibly quoted, escaped representation of s.
func Escape(s string) string {
	l := len(s)

	if l == 0 {
		return `""`
	}

	buf, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}

	j := string(buf)

	// if the json marshaling only added quotes, and the string does not contain
	// any "special" characters, use the original string.
	if len(j) == l+2 && !strings.ContainsAny(s, ` =@!:(){}`) {
		return s
	}

	return j
}
