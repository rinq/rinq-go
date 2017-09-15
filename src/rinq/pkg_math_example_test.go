// +build !without_amqp,!without_examples

package rinq_test

import (
	"context"
	"fmt"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/amqp"
)

// arguments contains the parameters for the commands in the "math" namespace
type arguments struct {
	Left, Right int
}

// mathHandler is the command handler for the "math" namespace
func mathHandler(
	ctx context.Context,
	req rinq.Request,
	res rinq.Response,
) {
	defer req.Payload.Close()

	// decode the request payload into the arguments struct
	var args arguments
	if err := req.Payload.Decode(&args); err != nil {
		res.Fail("invalid-arguments", "could not decode arguments")
		return
	}

	var result int

	switch req.Command {
	case "add":
		result = args.Left + args.Right
	case "sub":
		result = args.Left - args.Right
	default:
		res.Fail("unknown-command", "no such command: "+req.Command)
		return
	}

	// send the result in the response payload
	payload := rinq.NewPayload(result)
	defer payload.Close()

	res.Done(payload)
}

// This example shows how to issue a command call from one peer to another.
//
// There is a "server" peer, which performs basic mathematical operations,
// and a "client" peer which invokes those operations.
//
// In the example both the client peer and the server peer are running in the
// same process. Outside of an example, these peers would typically be running
// on separate servers.
func Example_mathService() {
	// create a new peer to act as the "server" and start listening for commands
	// in the "math" namespace.
	serverPeer, err := amqp.DialEnv()
	if err != nil {
		panic(err)
	}
	defer serverPeer.Stop()
	serverPeer.Listen("math", mathHandler)

	// create a new peer to act as the "client", and a session to make the
	// call.
	clientPeer, err := amqp.DialEnv()
	if err != nil {
		panic(err)
	}
	defer clientPeer.Stop()

	sess := clientPeer.Session()
	defer sess.Destroy()

	// call the "math::add" command
	ctx := context.Background()
	args := rinq.NewPayload(arguments{1, 2})
	result, err := sess.Call(ctx, "math", "add", args)
	if err != nil {
		panic(err)
	}

	fmt.Printf("1 + 2 = %s\n", result)
	// Output: 1 + 2 = 3
}
