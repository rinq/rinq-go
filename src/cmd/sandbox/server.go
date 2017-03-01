package main

import (
	"context"

	"github.com/rinq/rinq-go/src/rinq"
)

func runServer(peer rinq.Peer) {
	peer.Listen("our-namespace", func(
		ctx context.Context,
		req rinq.Request,
		res rinq.Response,
	) {
		defer req.Payload.Close()

		_, err := req.Source.Update(ctx, rinq.Set("foo", "bar"))
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
