package nsutil

import (
	"errors"
	"fmt"
	"regexp"
)

// Validate checks if ns is a valid namespace.
//
// Namespaces must not be empty. Valid characters are alpha-numeric characters,
// underscores, hyphens, periods and colons.
//
// Namespaces beginning with an underscore are reserved for internal use.
//
// The return value is nil if ns is a valid, unreserved namespace.
func Validate(ns string) error {
	if ns == "" {
		return errors.New("namespace must not be empty")
	} else if ns[0] == '_' {
		return fmt.Errorf("namespace '%s' is reserved", ns)
	} else if !namespacePattern.MatchString(ns) {
		return fmt.Errorf("namespace '%s' contains invalid characters", ns)
	}

	return nil
}

var namespacePattern *regexp.Regexp

func init() {
	namespacePattern = regexp.MustCompile(`^[A-Za-z0-9_\.\-:]+$`)
}
