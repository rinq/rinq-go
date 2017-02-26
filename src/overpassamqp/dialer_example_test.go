// +build !nobroker

package overpassamqp_test

import (
	"fmt"

	"github.com/over-pass/overpass-go/src/overpassamqp"
)

func ExampleDial() {
	_, err := overpassamqp.Dial("amqp://localhost")
	if err != nil {
		panic(err)
	}

	fmt.Println("connected")
	// Output: connected
}
