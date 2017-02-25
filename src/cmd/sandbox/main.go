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

	stopping := false

	for {
		select {
		case sig := <-signals:
			if stopping {
				// TODO: reimplement everything using service.StateMachine
				fmt.Println(" -- forceful stop:", sig)
				go peer.Stop()
			} else {
				stopping = true
				fmt.Println(" -- graceful stop:", sig)
				go peer.GracefulStop()
			}
		case <-peer.Done():
			if err := peer.Err(); err != nil {
				panic(err)
			}

			return
		}
	}
}
