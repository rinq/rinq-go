package main

import (
	"context"
	"fmt"
	"time"

	"github.com/rinq/rinq-go/src/rinq"
)

func runListener(peer rinq.Peer) {
	sess := peer.Session()
	defer sess.Destroy()

	sess.Listen(handle)

	<-sess.Done()
}

func handle(
	ctx context.Context,
	target rinq.Session,
	n rinq.Notification,
) {
	defer n.Payload.Close()

	fmt.Println("begin")
	time.Sleep(5 * time.Second)
	fmt.Println("end")
}
