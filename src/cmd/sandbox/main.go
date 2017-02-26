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

	peer, err := overpassamqp.DialConfig(
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

	switch os.Getenv("SANDBOX_ROLE") {
	case "server":
		go runServer(peer)
	case "notifier":
		go runNotifier(peer)
	case "listener":
		go runListener(peer)
	default:
		go runClient(peer)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	stopping := false

	for {
		select {
		case sig := <-signals:
			if stopping {
				fmt.Println(" -- forceful stop:", sig)
				peer.Stop()
			} else {
				stopping = true
				fmt.Println(" -- graceful stop:", sig)
				peer.GracefulStop()
			}
		case <-peer.Done():
			if err := peer.Err(); err != nil {
				panic(err)
			}

			return
		}
	}
}
