package strutil

import (
	"encoding/json"
	"regexp"
)

// Escape returns human-readable, possibly quoted, escaped representation of s.
func Escape(s string) string {
	if pattern.MatchString(s) {
		return s
	}

	buf, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}

	return string(buf)
}

var pattern *regexp.Regexp

func init() {
	pattern = regexp.MustCompile(`^[A-Za-z0-9_\.\-]+$`)
}
