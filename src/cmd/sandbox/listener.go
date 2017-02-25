package main

import (
	"context"
	"fmt"
	"time"

	"github.com/over-pass/overpass-go/src/overpass"
)

func runListener(peer overpass.Peer) {
	sess := peer.Session()
	defer sess.Close()

	sess.Listen(handle)

	<-sess.Done()
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
