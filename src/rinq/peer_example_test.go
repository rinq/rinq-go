// +build !without_amqp,!without_examples

package rinq_test

import (
	"context"
	"fmt"

	. "github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/amqp"
)

// This example illustrates how to establish a new session.
func ExamplePeer_session() {
	peer, err := amqp.Dial("")
	if err != nil {
		panic(err)
	}
	defer peer.Stop()

	sess := peer.Session()
	defer sess.Destroy()

	fmt.Printf("created session #%d\n", sess.ID().Seq)
	// Output: created session #1
}

// This example illustrates how to listen for incoming command requests.
func ExamplePeer_listen() {
	peer, err := amqp.Dial("")
	if err != nil {
		panic(err)
	}
	defer peer.Stop()

	peer.Listen("my-api", func(
		ctx context.Context,
		req Request,
		res Response,
	) {
		defer req.Payload.Close()
		// handle the command
		res.Close()
	})

	if false { // prevent the example from blocking forever.
		<-peer.Done()
	}
}
