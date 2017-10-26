package rinq

import "github.com/rinq/rinq-go/src/internal/namespaces"

// ValidateNamespace checks if ns is a valid namespace.
//
// Namespaces must not be empty. Valid characters are alpha-numeric characters,
// underscores, hyphens, periods and colons.
//
// Namespaces beginning with an underscore are reserved for internal use.
//
// The return value is nil if ns is a valid, unreserved namespace.
func ValidateNamespace(ns string) error {
	return namespaces.Validate(ns)
}
