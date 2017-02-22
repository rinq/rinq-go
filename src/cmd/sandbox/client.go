package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpassamqp"
)

func runClient() {
	rand.Seed(time.Now().UnixNano())

	peer, err := overpassamqp.Dial(
		context.Background(),
		"amqp://localhost",
		overpass.Config{
			Logger: overpass.NewLogger(true),
		},
	)
	if err != nil {
		panic(err)
	}
	defer peer.Stop()

	sess := peer.Session()
	defer sess.Close()

	for {
		func() {
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
			defer cancel()

			result, err := sess.Call(
				ctx,
				"our-namespace",
				"<whatever>",
				nil,
			)
			defer result.Close()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(result.Value())
			}
		}()
	}

	// fmt.Println(peer.Wait())
}
