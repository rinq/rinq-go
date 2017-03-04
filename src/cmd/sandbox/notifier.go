package main

import (
	"context"
	"fmt"
	"time"

	"github.com/rinq/rinq-go/src/rinq"
)

func runNotifier(peer rinq.Peer) {
	sess := peer.Session()
	defer sess.Destroy()

	go notify(sess)

	<-sess.Done()
}

func notify(sess rinq.Session) {
	for {
		time.Sleep(1 * time.Second)
		err := sess.NotifyMany(
			context.Background(),
			nil,
			"<whatever>",
			nil,
		)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}
