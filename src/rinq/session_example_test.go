package rinq_test

import (
	"context"
	"fmt"

	. "github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/amqp"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// This example shows how to make an asynchronous command call.
func ExampleSession_callAsync() {
	peer, err := amqp.Dial("")
	if err != nil {
		panic(err)
	}
	defer peer.Stop()

	// listen for command requests
	peer.Listen("my-api", func(
		ctx context.Context,
		req Request,
		res Response,
	) {
		defer req.Payload.Close()

		payload := NewPayload("<payload>")
		defer payload.Close()

		res.Done(payload)
	})

	sess := peer.Session()
	defer sess.Destroy()

	// setup the asynchronous response handler
	if err := sess.SetAsyncHandler(func(
		ctx context.Context,
		s Session, _ ident.MessageID,
		ns, cmd string,
		in *Payload, err error,
	) {
		defer in.Close()
		peer.Stop()

		fmt.Printf("received %s::%s response with %s payload\n", ns, cmd, in.Value())
	}); err != nil {
		panic(err)
	}

	// send the command request
	if _, err := sess.CallAsync(
		context.Background(),
		"my-api",
		"test",
		nil,
	); err != nil {
		panic(err)
	}

	<-peer.Done()
	// Output: received my-api::test response with <payload> payload
}

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
