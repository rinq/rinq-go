package env

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// UInt parses and validates a non-zero integer from the environment
// variable named v.
func UInt(v string) (uint, bool, error) {
	if s := os.Getenv(v); s != "" {
		n, err := strconv.ParseUint(s, 10, 31)
		if err != nil || n == 0 {
			return 0, false, fmt.Errorf("%s must be a non-zero integer", v)
		}

		return uint(n), true, nil
	}

	return 0, false, nil
}

// Duration parses and validates a non-zero duration in milliseconds
// from the environment variable named v.
func Duration(v string) (time.Duration, bool, error) {
	if s := os.Getenv(v); s != "" {
		n, err := strconv.ParseUint(s, 10, 63)
		if err != nil {
			return 0, false, fmt.Errorf("%s must be a non-zero duration (in milliseconds)", v)
		}

		return time.Duration(n) * time.Millisecond, true, nil
	}

	return 0, false, nil
}

// Bool parses and validates a boolean string from the environment variable
// named v.
func Bool(v string) (bool, bool, error) {
	switch os.Getenv(v) {
	case "true":
		return true, true, nil
	case "false":
		return false, true, nil
	case "":
		return false, false, nil
	default:
		return false, false, fmt.Errorf("%s must be 'true' or 'false'", v)
	}
}
