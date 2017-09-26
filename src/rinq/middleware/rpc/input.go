package rpc

import (
	"context"
	"reflect"

	"github.com/rinq/rinq-go/src/rinq"
)

var (
	contextType  = reflect.TypeOf((*context.Context)(nil)).Elem()
	revisionType = reflect.TypeOf((*rinq.Revision)(nil)).Elem()
)

type inputDecoder func(context.Context, rinq.Request) ([]reflect.Value, error)

// makeInputDecoder returns a function that produces a slice of reflect.Value
// suitable for passing as input to fn.
func makeInputDecoder(fn reflect.Type) inputDecoder {
	arity := fn.NumIn()
	decoders := make([]argDecoder, arity)
	foundParam := false

	for i := 0; i < arity; i++ {
		arg := fn.In(i)

		if arg.AssignableTo(contextType) {
			decoders[i] = contextDecoder
		} else if arg.AssignableTo(revisionType) {
			decoders[i] = revisionDecoder
		} else if foundParam {
			panic("function accepts too many input parameters")
		} else {
			decoders[i] = extractorForType(arg)
			foundParam = true
		}
	}

	return func(ctx context.Context, req rinq.Request) ([]reflect.Value, error) {
		in := make([]reflect.Value, arity)
		var err error

		for i, e := range decoders {
			in[i], err = e(ctx, req)

			if err != nil {
				return nil, err
			}
		}

		return in, nil
	}
}

type argDecoder func(context.Context, rinq.Request) (reflect.Value, error)

func contextDecoder(ctx context.Context, _ rinq.Request) (reflect.Value, error) {
	return reflect.ValueOf(ctx), nil
}

func revisionDecoder(_ context.Context, req rinq.Request) (reflect.Value, error) {
	return reflect.ValueOf(req.Source), nil
}

func extractorForType(t reflect.Type) argDecoder {
	return func(_ context.Context, req rinq.Request) (reflect.Value, error) {
		val := reflect.New(t)

		if err := req.Payload.Decode(val.Interface()); err != nil {
			return val, err
		}

		return val.Elem(), nil
	}
}
