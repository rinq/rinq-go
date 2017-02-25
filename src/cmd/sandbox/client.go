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

	// for i := 0; i < 4; i++ {
	// 	go call(sess)
	// }
	//
	// call(sess)

	sess.Listen(handle)

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

func handle(
	ctx context.Context,
	target overpass.Session,
	n overpass.Notification,
) {
	defer n.Payload.Close()

	fmt.Println("begin")
	time.Sleep(5 * time.Second)
	fmt.Println("end")
}
