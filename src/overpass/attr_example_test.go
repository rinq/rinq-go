// +build !nobroker

package overpass_test

import (
	"fmt"

	. "github.com/over-pass/overpass-go/src/overpass"
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
