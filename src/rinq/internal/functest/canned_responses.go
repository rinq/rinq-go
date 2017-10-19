package functest

import (
	"context"
	"time"

	"github.com/rinq/rinq-go/src/rinq"
)

// AlwaysReturn creates a CommandHandler that already responds with v.
func AlwaysReturn(v interface{}) rinq.CommandHandler {
	p := rinq.NewPayload(v)

	return func(ctx context.Context, req rinq.Request, res rinq.Response) {
		req.Payload.Close()
		res.Done(p)
	}
}

// AlwaysPanic creates a CommandHandler that already responds with v.
func AlwaysPanic() rinq.CommandHandler {
	return func(ctx context.Context, req rinq.Request, res rinq.Response) {
		req.Payload.Close()
		res.Close()
		panic("functest.AlwaysPanic!")
	}
}

// CloseAfter creates a CommandHandler that closes the response after a timeout.
func CloseAfter(d time.Duration) rinq.CommandHandler {
	return func(ctx context.Context, req rinq.Request, res rinq.Response) {
		req.Payload.Close()
		time.Sleep(d)
		res.Close()
	}
}

// Barrier create a command handler that attempts to write to ch twice.
func Barrier(ch chan<- struct{}) rinq.CommandHandler {
	return BarrierN(ch, 2)
}

// BarrierN create a command handler that attempts to write to ch n times.
func BarrierN(ch chan<- struct{}, n int) rinq.CommandHandler {
	return func(ctx context.Context, req rinq.Request, res rinq.Response) {
		req.Payload.Close()
		defer res.Close()

		for i := 0; i < n; i++ {
			select {
			case ch <- struct{}{}:
			case <-ctx.Done():
				return
			}
		}
	}
}
