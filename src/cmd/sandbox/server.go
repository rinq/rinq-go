package main

import (
	"context"
	"time"

	"github.com/over-pass/overpass-go/src/overpass"
)

func runServer(peer overpass.Peer) {
	peer.Listen("our-namespace", func(
		ctx context.Context,
		cmd overpass.Command,
		res overpass.Responder,
	) {
		defer cmd.Payload.Close()
		time.Sleep(5 * time.Second)
		res.Close()
	})

	<-peer.Done()
}
