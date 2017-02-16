package main

import (
	"context"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/over-pass/overpass-go/src/overpass"
)

func runClient(peer overpass.Peer) error {
	sess := peer.Session()
	defer sess.Close()

	// if err := sess.Listen(onNotify); err != nil {
	// 	return err
	// }

	rev, err := sess.CurrentRevision()
	if err != nil {
		return err
	}

	rev, err = rev.Update(
		context.Background(),
		overpass.Freeze("product", "myapp"),
		overpass.Set("accountId", "7"),
		overpass.Set("foo", "bar"),
	)
	if err != nil {
		return err
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	res, err := sess.Call(
		ctx,
		"myapp.v1",
		"purchase",
		overpass.NewPayload("hello, joe"),
	)
	defer res.Close()
	if err != nil {
		spew.Dump(err)
	}

	// if err = sess.NotifyMany(
	// 	context.Background(),
	// 	overpass.Constraint{
	// 		"product":   "myapp",
	// 		"accountId": "7",
	// 	},
	// 	"initial-multicast",
	// 	nil,
	// ); err != nil {
	// 	return err
	// }

	// <-sess.Done()
	return nil
}

// func onNotify(
// 	ctx context.Context,
// 	sess overpass.Session,
// 	n overpass.Notification,
// ) {
// 	defer n.Payload.Close()
//
// 	if n.Type == "initial-multicast" {
// 		if err := sess.Notify(
// 			ctx,
// 			n.Source.Ref().ID,
// 			"reply",
// 			overpass.NewPayload(1),
// 		); err != nil {
// 			fmt.Println(err)
// 		}
// 	} else if n.Type == "reply" {
// 		var counter uint
// 		n.Payload.Decode(&counter)
//
// 		if counter == 3 {
// 			_, err := n.Source.Update(
// 				ctx,
// 				overpass.Set("foo", ""),
// 				overpass.Freeze("accountId", ""),
// 			)
// 			if err != nil {
// 				fmt.Println(err)
// 			}
//
// 			sess.Close()
//
// 			return
// 		}
//
// 		if err := sess.Notify(
// 			ctx,
// 			n.Source.Ref().ID,
// 			"reply",
// 			overpass.NewPayload(counter+1),
// 		); err != nil {
// 			fmt.Println(err)
// 		}
// 	}
// }
