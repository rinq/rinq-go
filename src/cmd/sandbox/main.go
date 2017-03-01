package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/amqp"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	peer, err := amqp.DialConfig(
		context.Background(),
		os.Getenv("AMQP_DSN"),
		rinq.Config{
			Logger:        rinq.NewLogger(true),
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
		case <-signals:
			if stopping {
				fmt.Printf("\n-- forceful stop --\n\n")
				peer.Stop()
			} else {
				stopping = true
				fmt.Printf("\n-- graceful stop --\n\n")
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
