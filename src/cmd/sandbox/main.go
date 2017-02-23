package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpassamqp"
)

func main() {
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

	if os.Getenv("OVERPASS_SERVER") != "" {
		go runServer(peer)
	} else {
		go runClient(peer)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	select {
	case sig := <-signals:
		fmt.Println("received signal: ", sig)
		peer.GracefulStop()
		<-peer.Done()
	case <-peer.Done():
	}

	if err := peer.Err(); err != nil {
		panic(err)
	}
}
