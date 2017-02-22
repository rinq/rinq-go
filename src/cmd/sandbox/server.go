package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpassamqp"
)

func runServer() {
	rand.Seed(time.Now().UnixNano())

	peer, err := overpassamqp.Dial(
		context.Background(),
		"amqp://localhost",
		overpass.Config{
			Logger:        overpass.NewLogger(true),
			PruneInterval: 10 * time.Second,
		},
	)
	if err != nil {
		panic(err)
	}
	defer peer.Stop()

	peer.Listen("our-namespace", func(
		ctx context.Context,
		cmd overpass.Command,
		res overpass.Responder,
	) {
		defer cmd.Payload.Close()

		time.Sleep(250 * time.Millisecond)

		fmt.Println(ctx.Deadline())

		if !res.IsRequired() {
			fmt.Println("NOT REQUIRED!")
		}
		res.Close()
	})

	<-peer.Done()
	if err := peer.Err(); err != nil {
		panic(err)
	}
}
