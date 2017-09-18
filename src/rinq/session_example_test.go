// +build !without_amqp,!without_examples

package rinq_test

import (
	"context"
	"fmt"
	"sync/atomic"

	. "github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/amqp"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// This example shows how to make an asynchronous command call.
func ExampleSession_callAsync() {
	peer, err := amqp.DialEnv()
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
	peer, err := amqp.DialEnv()
	if err != nil {
		panic(err)
	}
	defer peer.Stop()

	// create a session to receive the notification
	recv := peer.Session()
	defer recv.Destroy()

	if err := recv.Listen(
		"my-api",
		func(
			ctx context.Context,
			target Session,
			n Notification,
		) {
			defer n.Payload.Close()
			peer.Stop()

			fmt.Printf("received %s::%s with %s payload\n", n.Namespace, n.Type, n.Payload.Value())
		},
	); err != nil {
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
		"my-api",
		"<type>",
		payload,
	); err != nil {
		panic(err)
	}

	<-peer.Done()
	// Output: received my-api::<type> with <payload> payload
}

// This example shows how to send a notification from one session to several
// sessions that contain specific attribute values.
func ExampleSession_notifyMany() {
	peer, err := amqp.DialEnv()
	if err != nil {
		panic(err)
	}

	// create three sessions for receiving notifications
	recv1 := peer.Session()
	recv2 := peer.Session()
	recv3 := peer.Session()

	// create a notification handler that stops the peer once TWO notifications
	// have been received
	var recvCount int32
	handler := func(ctx context.Context, target Session, n Notification) {
		defer n.Payload.Close()

		if target == recv3 {
			panic("message delivered to unexpected session")
		}

		fmt.Printf("received %s::%s with %s payload\n", n.Namespace, n.Type, n.Payload.Value())

		if atomic.AddInt32(&recvCount, 1) == 2 {
			peer.Stop()
		}
	}

	// configure all three sessions to listen for notifications
	for _, s := range []Session{recv1, recv2, recv3} {
		if err := s.Listen("my-api", handler); err != nil {
			panic(err)
		}
	}

	// update the first TWO sessions with a "foo" attribute
	for _, s := range []Session{recv1, recv2} {
		rev, err := s.CurrentRevision()
		if err != nil {
			panic(err)
		}

		if _, err := rev.Update(context.Background(), Freeze("foo", "bar")); err != nil {
			panic(err)
		}
	}

	// create a session to send the notification to recv
	send := peer.Session()

	payload := NewPayload("<payload>")
	defer payload.Close()

	// constraint the notification to only those sessions that have a "foo"
	// attribute with a value of "bar"
	con := Constraint{
		"foo": "bar",
	}

	if err := send.NotifyMany(
		context.Background(),
		con,
		"my-api",
		"<type>",
		payload,
	); err != nil {
		panic(err)
	}

	<-peer.Done()
	// Output: received my-api::<type> with <payload> payload
	// received my-api::<type> with <payload> payload
}
