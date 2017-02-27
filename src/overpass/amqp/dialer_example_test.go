// +build !nobroker

package amqp_test

import (
	"context"
	"fmt"
	"time"

	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpass/amqp"
)

// This example demonstrates how to establish a peer on an Overpass network
// using the default Overpass configuration.
func ExampleDial() {
	peer, err := amqp.Dial("amqp://localhost")
	if err != nil {
		panic(err)
	}
	defer peer.Stop()

	fmt.Printf("connected")
	// Output: connected
}

// This example demonstrates how to establish a peer on an Overpass network
// using a custom Overpass configuration.
func ExampleDialConfig() {
	cfg := overpass.Config{
		DefaultTimeout: 10 * time.Second,
	}

	peer, err := amqp.DialConfig(context.Background(), "amqp://localhost", cfg)
	if err != nil {
		panic(err)
	}
	defer peer.Stop()

	fmt.Println("connected")
	// Output: connected
}

// This example demonstrates how to establish a peer on an Overpass network
// using a Dialer with a customer AMQP configuration.
func ExampleDialer() {
	dialer := &amqp.Dialer{}
	dialer.AMQPConfig.Heartbeat = 1 * time.Minute

	peer, err := dialer.Dial(
		context.Background(),
		"amqp://localhost",
		overpass.DefaultConfig,
	)
	if err != nil {
		panic(err)
	}
	defer peer.Stop()

	fmt.Println("connected")
	// Output: connected
}
