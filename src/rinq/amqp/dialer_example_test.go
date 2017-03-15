// +build !without_amqp,!without_examples

package amqp_test

import (
	"context"
	"fmt"
	"time"

	"github.com/rinq/rinq-go/src/rinq"
	. "github.com/rinq/rinq-go/src/rinq/amqp"
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
// a custom configuration.
func ExampleDialConfig() {
	cfg := rinq.Config{
		DefaultTimeout: 10 * time.Second,
	}

	peer, err := DialConfig(context.Background(), "amqp://localhost", cfg)
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
		rinq.DefaultConfig,
	)
	if err != nil {
		panic(err)
	}
	defer peer.Stop()

	fmt.Println("connected")
	// Output: connected
}
