package env

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Int parses and validates a non-zero integer from the environment
// variable named v, or returns d if v is undefined.
func Int(v string, d int) (int, error) {
	if s := os.Getenv(v); s != "" {
		n, err := strconv.ParseUint(s, 10, 31)
		if err != nil || n == 0 {
			return 0, fmt.Errorf("%s must be a non-zero integer", v)
		}

		return int(n), nil
	}

	return d, nil
}

// Duration parses and validates a non-zero duration in milliseconds
// from the environment variable named v, or returns d if v is undefined.
func Duration(v string, d time.Duration) (time.Duration, error) {
	if s := os.Getenv(v); s != "" {
		n, err := strconv.ParseUint(s, 10, 63)
		if err != nil {
			return 0, fmt.Errorf("%s must be a non-zero duration (in milliseconds)", v)
		}

		return time.Duration(n) * time.Millisecond, nil
	}

	return d, nil
}

// Bool parses and validates a boolean string from the environment variable
// named v, or returns d if v is undefined.
func Bool(v string, d bool) (bool, error) {
	switch os.Getenv(v) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "":
		return d, nil
	default:
		return false, fmt.Errorf("%s must be 'true' or 'false'", v)
	}
}
