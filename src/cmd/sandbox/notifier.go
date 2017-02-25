package main

import (
	"context"
	"fmt"
	"time"

	"github.com/over-pass/overpass-go/src/overpass"
)

func runNotifier(peer overpass.Peer) {
	sess := peer.Session()
	defer sess.Close()

	go notify(sess)

	<-sess.Done()
}

func notify(sess overpass.Session) {
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
