package main

import (
	"context"
	"fmt"
	"time"

	"github.com/over-pass/overpass-go/src/overpass"
)

func runClient(peer overpass.Peer) {
	sess := peer.Session()
	defer sess.Close()

	callAsync(sess)
	// sess.Close()

	<-sess.Done()
}

func call(sess overpass.Session) {
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

func callAsync(sess overpass.Session) {
	err := sess.SetAsyncHandler(func(
		ctx context.Context,
		msgID overpass.MessageID,
		ns string,
		cmd string,
		in *overpass.Payload,
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
