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

	for i := 0; i < 4; i++ {
		go call(sess)
	}

	// call(sess)

	<-sess.Done()
}

func call(sess overpass.Session) {
	for {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, 11*time.Second)
		defer cancel()

		result, err := sess.Call(
			ctx,
			"our-namespace",
			"<whatever>",
			nil,
		)
		defer result.Close()
		if err != nil {
			fmt.Println(err)
			break
		}
	}

	sess.Close()
}
