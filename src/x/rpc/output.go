package rpc

import (
	"reflect"

	"github.com/rinq/rinq-go/src/rinq"
)

var (
	errorType = reflect.TypeOf((*error)(nil)).Elem()
)

type outputEncoder func(out []reflect.Value) (*rinq.Payload, error)

func makeOutputEncoder(fn reflect.Type) outputEncoder {
	arity := fn.NumOut()

	switch arity {
	case 0:
		return emptyEncoder

	case 1:
		p := fn.Out(0)

		if p.AssignableTo(errorType) {
			return errorEncoder
		}

		return valueEncoder

	case 2:
		p := fn.Out(1)

		if !p.AssignableTo(errorType) {
			panic("second output parameter must be an error")
		}

		return valueOrErrorEncoder
	}

	panic("function returns too many output parameters")
}

func emptyEncoder(out []reflect.Value) (*rinq.Payload, error) {
	return nil, nil
}

func valueEncoder(out []reflect.Value) (*rinq.Payload, error) {
	return rinq.NewPayload(out[0].Interface()), nil
}

func errorEncoder(out []reflect.Value) (*rinq.Payload, error) {
	err := out[0]

	if err.IsNil() {
		return nil, nil
	}

	return nil, err.Interface().(error)
}

func valueOrErrorEncoder(out []reflect.Value) (*rinq.Payload, error) {
	val := out[0]
	err := out[1]

	if !err.IsNil() {
		return nil, err.Interface().(error)
	}

	return rinq.NewPayload(val.Interface()), nil
}
