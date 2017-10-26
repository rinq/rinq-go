package commandamqp

import "github.com/rinq/rinq-go/src/rinq"

// debugResponse wraps are "parent" response and captures the payload and error.
type debugResponse struct {
	res rinq.Response

	Payload *rinq.Payload
	Err     error
}

func newDebugResponse(parent rinq.Response) rinq.Response {
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

func (r *debugResponse) Done(payload *rinq.Payload) {
	r.res.Done(payload)
	r.Payload = payload.Clone()
}

func (r *debugResponse) Error(err error) {
	r.res.Error(err)
	r.Err = err
	if failure, ok := err.(rinq.Failure); ok {
		r.Payload = failure.Payload.Clone()
	}
}

func (r *debugResponse) Fail(t, f string, v ...interface{}) rinq.Failure {
	err := r.res.Fail(t, f, v...)
	r.Err = err
	return err
}

func (r *debugResponse) Close() bool {
	return r.res.Close()
}
