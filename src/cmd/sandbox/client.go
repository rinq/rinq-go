package main

import (
	"context"
	"fmt"
	"time"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

func runClient(peer rinq.Peer) {
	sess := peer.Session()
	defer sess.Destroy()

	callAsync(sess)
	// sess.Destroy()

	<-sess.Done()
}

func call(sess rinq.Session) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 11*time.Second)
	defer cancel()

	_, err := sess.Call(
		ctx,
		"our-namespace",
		"<whatever>",
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}
}

func callAsync(sess rinq.Session) {
	err := sess.SetAsyncHandler(func(
		ctx context.Context,
		msgID ident.MessageID,
		ns string,
		cmd string,
		in *rinq.Payload,
		err error,
	) {
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(ns, cmd, in, err)
	})

	if err != nil {
		fmt.Println(err)
	}

	_, err = sess.CallAsync(
		context.Background(),
		"our-namespace",
		"<whatever>",
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}
}
