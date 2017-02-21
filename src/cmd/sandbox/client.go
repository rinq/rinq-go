package main

import (
	"context"

	"github.com/over-pass/overpass-go/src/overpass"
)

func runClient(peer overpass.Peer) error {
	sess := peer.Session()
	defer sess.Close()

	if err := sess.Listen(onNotify); err != nil {
		return err
	}

	rev, err := sess.CurrentRevision()
	if err != nil {
		return err
	}

	rev, err = rev.Update(
		context.Background(),
		overpass.Freeze("product", "myapp"),
	)
	if err != nil {
		return err
	}

	response, err := sess.Call(
		context.Background(),
		"myapp.v1",
		"authenticate",
		nil,
	)
	defer response.Close()
	if err != nil {
		return err
	}

	if err = sess.NotifyMany(
		context.Background(),
		overpass.Constraint{
			"product":   "myapp",
			"accountId": "7",
		},
		"account-update",
		nil,
	); err != nil {
		return err
	}

	<-sess.Done()
	return nil
}

func onNotify(
	ctx context.Context,
	sess overpass.Session,
	n overpass.Notification,
) {
	defer n.Payload.Close()
}
