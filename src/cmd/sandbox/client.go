package main

import (
	"context"
	"fmt"

	"github.com/over-pass/overpass-go/src/overpass"
)

func runClient(peer overpass.Peer) error {
	sess := peer.Session()
	defer sess.Close()

	// if err := sess.Listen(onNotify); err != nil {
	// 	return err
	// }

	res, err := sess.Call(
		context.Background(),
		"myapp.v1",
		"command-name",
		nil,
	)
	defer res.Close()
	if err != nil {
		return err
	}

	rev, err := sess.CurrentRevision()
	if err != nil {
		return err
	}

	fmt.Println(rev.Get(context.Background(), "product"))

	// rev, err = rev.Update(
	// 	context.Background(),
	// 	overpass.Freeze("product", "myapp"),
	// 	overpass.Set("accountId", "7"),
	// 	overpass.Set("foo", "bar"),
	// )
	// if err != nil {
	// 	return err
	// }

	// var counter uint64
	// for {
	// 	counter++
	//
	// 	ctx := context.Background()
	//
	// 	rev, err = rev.Update(
	// 		ctx,
	// 		overpass.Set("counter", strconv.FormatUint(counter, 10)),
	// 	)
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	// 	defer cancel()
	//
	// 	res, err := sess.Call(
	// 		ctx,
	// 		"myapp.v1",
	// 		"read-counter",
	// 		nil,
	// 	)
	// 	defer res.Close()
	// 	if err != nil {
	// 		spew.Dump(err)
	// 		break
	// 	}
	//
	// 	time.Sleep(time.Second * 10)
	// }

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

	<-sess.Done()
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
