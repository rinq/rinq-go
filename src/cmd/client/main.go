package main

import (
	"context"
	"fmt"
	"math/rand"
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
			Logger: overpass.NewLogger(true),
		},
	)
	if err != nil {
		panic(err)
	}
	defer peer.Close()

	sess := peer.Session()
	defer sess.Close()

	ctx := context.Background()
	// ctx, cancel := context.WithTimeout(
	// 	context.Background(),
	// 	5000*time.Millisecond,
	// )
	// defer cancel()

	go auth(ctx, sess, "ABCD")
	go auth(ctx, sess, "EF73")

	<-sess.Done()
}

func auth(ctx context.Context, sess overpass.Session, ticket string) {
	payload := overpass.NewPayload(ticket)
	defer payload.Close()

	result, err := sess.Call(
		ctx,
		"auth.v1",
		"authByTicket",
		payload,
	)
	defer result.Close()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(result.Value())
	}
}
