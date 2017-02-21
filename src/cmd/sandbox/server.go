package main

import (
	"context"
	"fmt"

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

			rev, err := cmd.Source.Update(ctx, overpass.Set("product", "myapp"))
			if err != nil {
				fmt.Println("update error:", err)
				res.Error(err)
				return
			}

			err = rev.Close(ctx)
			if err != nil {
				fmt.Println("close error:", err)
				res.Error(err)
				return
			}

			res.Fail("invalid-widget", "it all went so wrong")

			// attr, err := cmd.Source.Get(ctx, "counter")
			// if err != nil {
			// 	fmt.Println("get() error:", err)
			// 	res.Error(err)
			// } else {
			// 	res.Close()
			// }
			//
			// spew.Dump(attr)
			// _, err := cmd.Source.Update(ctx, overpass.Set("foo", "bar"))
			// if err != nil {
			// 	res.Error(err)
			// 	return
			// }

			// res.Fail("insufficient-funds", "account 7 is broke!")
		},
	); err != nil {
		panic(err)
	}

	return peer.Wait()
}
