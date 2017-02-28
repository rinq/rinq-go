package main

import (
	"context"

	"github.com/over-pass/overpass-go/src/overpass"
)

func runServer(peer overpass.Peer) {
	peer.Listen("our-namespace", func(
		ctx context.Context,
		req overpass.Request,
		res overpass.Response,
	) {
		defer req.Payload.Close()

		_, err := req.Source.Update(ctx, overpass.Set("foo", "bar"))
		if err != nil {
			res.Fail("cant-update", "failed to set attributes on the source session")
			return
		}

		// err = rev.Close(ctx)
		// if err != nil {
		// 	res.Fail("cant-close", "failed to close the source session")
		// 	return
		// }

		res.Close()
	})

	<-peer.Done()
}
