package command

import (
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/internal/opentr"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// response wraps a "parent" response and performs logging and tracing when the
// response is closed.
type response struct {
	req rinq.Request
	res rinq.Response

	peerID    ident.PeerID
	traceID   string
	logger    rinq.Logger
	span      opentracing.Span
	startedAt time.Time
}

// NewResponse returns a response that wraps res.
func NewResponse(
	req rinq.Request,
	res rinq.Response,
	peerID ident.PeerID,
	traceID string,
	logger rinq.Logger,
	span opentracing.Span,
) rinq.Response {
	return &response{
		res: res,
		req: req,

		peerID:    peerID,
		traceID:   traceID,
		logger:    logger,
		span:      span,
		startedAt: time.Now(),
	}
}

func (r *response) IsRequired() bool {
	return r.res.IsRequired()
}

func (r *response) IsClosed() bool {
	return r.res.IsClosed()
}

func (r *response) Done(payload *rinq.Payload) {
	r.res.Done(payload)
	r.logSuccess(payload)

	opentr.LogServerSuccess(r.span, payload)
}

func (r *response) Error(err error) {
	r.res.Error(err)

	if failure, ok := err.(rinq.Failure); ok {
		r.logFailure(failure.Type, failure.Payload)
	} else {
		r.logError(err)
	}

	opentr.LogServerError(r.span, err)
}

func (r *response) Fail(f, t string, v ...interface{}) rinq.Failure {
	err := r.res.Fail(f, t, v...)
	r.logFailure(f, nil)
	opentr.LogServerError(r.span, err)

	return err
}

func (r *response) Close() bool {
	if r.res.Close() {
		r.logSuccess(nil)
		opentr.LogServerSuccess(r.span, nil)
		return true
	}

	return false
}

func (r *response) logSuccess(payload *rinq.Payload) {
	r.logger.Log(
		"%s handled '%s::%s' command from %s successfully (%dms %d/i %d/o) [%s]",
		r.peerID.ShortString(),
		r.req.Namespace,
		r.req.Command,
		r.req.ID.Ref.ShortString(),
		time.Since(r.startedAt)/time.Millisecond,
		r.req.Payload.Len(),
		payload.Len(),
		r.traceID,
	)
}

func (r *response) logFailure(failureType string, payload *rinq.Payload) {
	r.logger.Log(
		"%s handled '%s::%s' command from %s: '%s' failure (%dms %d/i %d/o) [%s]",
		r.peerID.ShortString(),
		r.req.Namespace,
		r.req.Command,
		r.req.ID.Ref.ShortString(),
		failureType,
		time.Since(r.startedAt)/time.Millisecond,
		r.req.Payload.Len(),
		payload.Len(),
		r.traceID,
	)
}

func (r *response) logError(err error) {
	r.logger.Log(
		"%s handled '%s::%s' command from %s: '%s' error (%dms %d/i 0/o) [%s]",
		r.peerID.ShortString(),
		r.req.Namespace,
		r.req.Command,
		r.req.ID.Ref.ShortString(),
		err,
		time.Since(r.startedAt)/time.Millisecond,
		r.req.Payload.Len(),
		r.traceID,
	)
}
