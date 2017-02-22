package main

import (
	"context"
	"fmt"
	"time"

	"github.com/over-pass/overpass-go/src/overpass"
)

func runServer(peer overpass.Peer) error {
	if err := peer.Listen(
		"myapp.v1",
		func(
			ctx context.Context,
			cmd overpass.Command,
			res overpass.Responder,
		) {
			defer cmd.Payload.Close()

			rev := cmd.Source
			var err error

			for {
				rev, err = rev.Update(
					ctx,
					overpass.Freeze("accountId", "7"),
				)
				if err != nil {
					if overpass.ShouldRetry(err) {
						rev, err = rev.Refresh(ctx)
						if err == nil {
							continue
						}
					}

					res.Error(err)
					return
				}

				break
			}

			res.Close()
			// res.Error(overpass.Failure{
			// 	Type:    "invalid-widget",
			// 	Message: "it all went so wrong",
			// 	Payload: overpass.NewPayload("OH SNAP"),
			// })

			go func() {
				fmt.Println("sleeping")
				time.Sleep(10 * time.Second)
				ctx := context.Background()

				for {
					fmt.Println("closing")

					if err := rev.Close(ctx); err != nil {
						if overpass.ShouldRetry(err) {
							fmt.Println("refreshing")
							rev, err = rev.Refresh(ctx)
							if err == nil {
								continue
							}
						}

						fmt.Println(err)
					}

					return
				}
			}()
		},
	); err != nil {
		panic(err)
	}

	<-peer.Done()
	return peer.Err()
}
