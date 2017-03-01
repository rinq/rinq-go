// +build !nobroker

package rinq_test

import (
	"fmt"

	. "github.com/rinq/rinq-go/src/rinq"
)

func ExampleSet() {
	attr := Set("foo", "bar")

	fmt.Println(attr)
	// Output: foo=bar
}

func ExampleFreeze() {
	attr := Freeze("foo", "bar")

	fmt.Println(attr)
	// Output: foo@bar
}
