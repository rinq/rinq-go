package rinq_test

import (
	"context"
	"fmt"

	. "github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/amqp"
)

// This example shows how to send a notification from one session to another.
func ExampleSession_notify() {
	peer, err := amqp.Dial("")
	if err != nil {
		panic(err)
	}
	defer peer.Stop()

	// create a session to receive the notification
	recv := peer.Session()
	defer recv.Destroy()

	if err := recv.Listen(func(
		ctx context.Context,
		target Session,
		n Notification,
	) {
		defer n.Payload.Close()
		peer.Stop()

		fmt.Printf("received %s with %s payload\n", n.Type, n.Payload.Value())
	}); err != nil {
		panic(err)
	}

	// create a session to send the notification to recv
	send := peer.Session()
	defer send.Destroy()

	payload := NewPayload("<payload>")
	defer payload.Close()

	if err := send.Notify(
		context.Background(),
		recv.ID(),
		"<type>",
		payload,
	); err != nil {
		panic(err)
	}

	<-peer.Done()
	// Output: received <type> with <payload> payload
}
