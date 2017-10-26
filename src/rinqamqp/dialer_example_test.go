// +build !without_amqp,!without_examples

package rinqamqp_test

import (
	"context"
	"fmt"
	"time"

	"github.com/rinq/rinq-go/src/rinq/options"
	. "github.com/rinq/rinq-go/src/rinqamqp"
)

// This example demonstrates how to establish a peer on a Rinq network using
// the default configuration.
func ExampleDial() {
	peer, err := Dial("amqp://localhost")
	if err != nil {
		panic(err)
	}
	defer peer.Stop()

	fmt.Printf("connected")
	// Output: connected
}

// This example demonstrates how to establish a peer on a Rinq network using
// custom options.
func ExampleDial_withOptions() {
	peer, err := Dial(
		"amqp://localhost",
		options.DefaultTimeout(10*time.Second),
	)
	if err != nil {
		panic(err)
	}
	defer peer.Stop()

	fmt.Println("connected")
	// Output: connected
}

// This example demonstrates how to establish a peer on a Rinq network using a
// Dialer with a custom AMQP configuration.
func ExampleDialer() {
	dialer := &Dialer{}
	dialer.AMQPConfig.Heartbeat = 1 * time.Minute

	peer, err := dialer.Dial(
		context.Background(),
		"amqp://localhost",
	)
	if err != nil {
		panic(err)
	}
	defer peer.Stop()

	fmt.Println("connected")
	// Output: connected
}
