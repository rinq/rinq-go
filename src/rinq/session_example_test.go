// +build !without_amqp,!without_examples

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
// sessions that contain a specific attribute value.
func ExampleSession_notifyMany() {
	peer, err := amqp.DialEnv()
	if err != nil {
		panic(err)
	}
	defer peer.Stop()

	// setup a command handler that sets an attribute on the source session
	if err := peer.Listen(
		"my-api",
		func(ctx context.Context, req Request, res Response) {
			if _, err := req.Source.Update(ctx, Freeze("foo", "bar")); err != nil {
				res.Error(err)
			}
			res.Close()
		},
	); err != nil {
		panic(err)
	}

	// create three sessions, two of which will match the constraint and hence
	// receive the notification
	//
	// note that all three sessions are listening, but only two of them invoke
	// the command above to set the "foo" attribute
	var recv []Session
	defer func() {
		for _, s := range recv {
			s.Destroy()
		}
	}()

	count := 0
	handler := func(ctx context.Context, target Session, n Notification) {
		defer n.Payload.Close()

		if target == recv[2] {
			panic("session received unexpected notification")
		}

		fmt.Printf("received %s::%s with %s payload\n", n.Namespace, n.Type, n.Payload.Value())

		count++

		if count == 2 {
			peer.Stop()
		}
	}

	for i := 0; i < 3; i++ {
		s := peer.Session()
		recv = append(recv, s)

		if err := s.Listen("my-api", handler); err != nil {
			panic(err)
		}

		if i < 2 {
			if _, err := s.Call(context.Background(), "my-api", "set-attr", nil); err != nil {
				panic(err)
			}
		}
	}

	// create a session to send the notification to recv
	send := peer.Session()
	defer send.Destroy()

	payload := NewPayload("<payload>")
	defer payload.Close()

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
