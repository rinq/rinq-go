package commandamqp

import "github.com/over-pass/overpass-go/src/overpass"

type capturingResponder struct {
	parent  overpass.Responder
	payload *overpass.Payload
	err     error
}

func (r *capturingResponder) Response() (*overpass.Payload, error) {
	return r.payload, r.err
}

func newCapturingResponder(
	parent overpass.Responder,
) overpass.Responder {
	return &capturingResponder{
		parent: parent,
	}
}

func (r *capturingResponder) IsRequired() bool {
	return r.parent.IsRequired()
}

func (r *capturingResponder) IsClosed() bool {
	return r.parent.IsClosed()
}

func (r *capturingResponder) Done(payload *overpass.Payload) {
	r.parent.Done(payload)
	r.payload = payload.Clone()
}

func (r *capturingResponder) Error(err error) {
	r.parent.Error(err)
	r.err = err
	if failure, ok := err.(overpass.Failure); ok {
		r.payload = failure.Payload.Clone()
	}
}

func (r *capturingResponder) Fail(failureType, message string) overpass.Failure {
	err := r.parent.Fail(failureType, message)
	r.err = err
	return err
}

func (r *capturingResponder) Close() bool {
	return r.parent.Close()
}
