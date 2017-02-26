package commandamqp

import "github.com/over-pass/overpass-go/src/overpass"

// debugResponse wraps are "parent" response and captures the payload and error.
type debugResponse struct {
	res overpass.Response

	Payload *overpass.Payload
	Err     error
}

func newDebugResponse(parent overpass.Response) overpass.Response {
	return &debugResponse{
		res: parent,
	}
}

func (r *debugResponse) IsRequired() bool {
	return r.res.IsRequired()
}

func (r *debugResponse) IsClosed() bool {
	return r.res.IsClosed()
}

func (r *debugResponse) Done(payload *overpass.Payload) {
	r.res.Done(payload)
	r.Payload = payload.Clone()
}

func (r *debugResponse) Error(err error) {
	r.res.Error(err)
	r.Err = err
	if failure, ok := err.(overpass.Failure); ok {
		r.Payload = failure.Payload.Clone()
	}
}

func (r *debugResponse) Fail(failureType, message string) overpass.Failure {
	err := r.res.Fail(failureType, message)
	r.Err = err
	return err
}

func (r *debugResponse) Close() bool {
	return r.res.Close()
}
