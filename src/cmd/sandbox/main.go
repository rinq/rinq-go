package main

import (
	"context"
	"math/rand"
	"os"
	"time"

	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpassamqp"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	peer, err := overpassamqp.Dial(
		context.Background(),
		"amqp://localhost",
		overpass.Config{},
	)
	if err != nil {
		panic(err)
	}
	defer peer.Close()

	if os.Getenv("OVERPASS_SERVER") != "" {
		err = runServer(peer)
	} else {
		err = runClient(peer)
	}

	if err != nil {
		panic(err)
	}
}
