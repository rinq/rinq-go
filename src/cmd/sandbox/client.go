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

	call(sess)
	sess.Close()

	<-sess.Done()
}

func call(sess overpass.Session) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 11*time.Second)
	defer cancel()

	err := sess.Execute(
		ctx,
		"our-namespace",
		"<whatever>",
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}
}
