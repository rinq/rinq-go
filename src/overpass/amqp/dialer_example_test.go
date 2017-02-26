// +build !nobroker

package amqp_test

import (
	"fmt"

	"github.com/over-pass/overpass-go/src/overpass/amqp"
)

func ExampleDial() {
	_, err := amqp.Dial("amqp://localhost")
	if err != nil {
		panic(err)
	}

	fmt.Println("connected")
	// Output: connected
}
