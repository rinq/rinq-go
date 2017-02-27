// +build !nobroker

package overpass_test

import (
	"context"
	"fmt"

	. "github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpass/amqp"
)

// This example illustrates how to establish a new session.
func ExamplePeer_session() {
	peer, err := amqp.Dial("")
	if err != nil {
		panic(err)
	}
	defer peer.Stop()

	sess := peer.Session()
	defer sess.Close()

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
