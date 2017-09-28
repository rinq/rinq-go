package rpc

import (
	"context"
	"reflect"

	"github.com/rinq/rinq-go/src/rinq"
)

// NewHandler takes an arbitrary function and returns a command handler that
// maps a command request to the input arguments and the output arguments to a
// response.
func NewHandler(fn interface{}) rinq.CommandHandler {
	t := reflect.TypeOf(fn)

	if t.Kind() != reflect.Func {
		panic("handler must be a function")
	}

	if t.IsVariadic() {
		panic("handler must not be variadic")
	}

	dec := makeInputDecoder(t)
	enc := makeOutputEncoder(t)

	v := reflect.ValueOf(fn)

	return func(ctx context.Context, req rinq.Request, res rinq.Response) {
		defer req.Payload.Close()

		in, err := dec(ctx, req)
		if err != nil {
			res.Error(err)
			return
		}

		out := v.Call(in)

		p, err := enc(out)
		defer p.Close()

		if err != nil {
			res.Error(err)
		} else {
			res.Done(p)
		}
	}
}
