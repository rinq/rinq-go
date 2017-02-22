package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
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

	peer.Listen("auth.v1", func(
		ctx context.Context,
		cmd overpass.Command,
		res overpass.Responder,
	) {
		defer cmd.Payload.Close()

		switch cmd.Command {
		case "authByTicket":
			authByTicket(ctx, cmd, res)
		default:
			res.Error(fmt.Errorf("unknown command: %s", cmd.Command))
		}
	})

	if err := peer.Wait(); err != nil {
		panic(err)
	}
}

func authByTicket(
	ctx context.Context,
	cmd overpass.Command,
	res overpass.Responder,
) {
	rev := cmd.Source

retry:
	customerID, err := rev.Get(ctx, "customerID")
	fmt.Println(customerID, err)
	if err != nil {
		ref := rev.Ref().ShortString()
		fmt.Printf("%s get: %s\n", ref, err)

		if overpass.ShouldRetry(err) {
			rev, err = rev.Refresh(ctx)
			if err == nil {
				goto retry
			}
			fmt.Printf("%s refresh: %s\n", ref, err)
		}

		if overpass.IsNotFound(err) {
			res.Close()
		} else {
			res.Error(errors.New("get failed"))
		}

		return
	}

	if customerID.Value != "" {
		res.Fail("already-authed", "you are already logged in")
		return
	}

	var ticket string

	if err = cmd.Payload.Decode(&ticket); err != nil {
		res.Fail("malformed-ticket", "tickets must be strings")
		return
	}

	ticketInt, err := strconv.ParseUint(ticket, 16, 64)
	if err != nil {
		res.Fail("malformed-ticket", "tickets must be hex")
		return
	}

	if ticketInt == 0 {
		res.Fail("invalid-ticket", "ticket could not be authenticated")
		return
	}

	cust := customer{ticketInt, "bob"}

	rev, err = rev.Update(
		ctx,
		overpass.Set("customerID", strconv.FormatUint(ticketInt, 10)),
	)
	if err != nil {
		ref := rev.Ref().ShortString()
		fmt.Printf("%s update: %s\n", ref, err)

		if overpass.ShouldRetry(err) {
			rev, err = rev.Refresh(ctx)
			if err == nil {
				goto retry
			}
			fmt.Printf("%s refresh: %s\n", ref, err)
		}

		if overpass.IsNotFound(err) {
			res.Close()
		} else {
			res.Error(errors.New("update failed"))
		}
		return
	}

	payload := overpass.NewPayload(cust)
	defer payload.Close()

	res.Done(payload)
}

type customer struct {
	ID       uint64
	Nickname string
}
