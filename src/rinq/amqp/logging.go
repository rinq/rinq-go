package amqp

import (
	"time"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

func logStartedListening(
	logger rinq.Logger,
	peerID ident.PeerID,
	namespace string,
) {
	logger.Log(
		"%s started listening for command requests in '%s' namespace",
		peerID.ShortString(),
		namespace,
	)
}

func logStoppedListening(
	logger rinq.Logger,
	peerID ident.PeerID,
	namespace string,
) {
	logger.Log(
		"%s stopped listening for command requests in '%s' namespace",
		peerID.ShortString(),
		namespace,
	)
}

// loggingResponse wraps are "parent" response and emits a log entry when it is
// closed.
type loggingResponse struct {
	req rinq.Request
	res rinq.Response

	peerID    ident.PeerID
	traceID   string
	logger    rinq.Logger
	startedAt time.Time
}

func newLoggingResponse(
	req rinq.Request,
	res rinq.Response,
	peerID ident.PeerID,
	traceID string,
	logger rinq.Logger,
) rinq.Response {
	return &loggingResponse{
		res: res,
		req: req,

		peerID:    peerID,
		traceID:   traceID,
		logger:    logger,
		startedAt: time.Now(),
	}
}

func (r *loggingResponse) IsRequired() bool {
	return r.res.IsRequired()
}

func (r *loggingResponse) IsClosed() bool {
	return r.res.IsClosed()
}

func (r *loggingResponse) Done(payload *rinq.Payload) {
	r.res.Done(payload)
	r.logSuccess(payload)
}

func (r *loggingResponse) Error(err error) {
	r.res.Error(err)

	if failure, ok := err.(rinq.Failure); ok {
		r.logFailure(failure.Type, failure.Payload)
	} else {
		r.logError(err)
	}
}

func (r *loggingResponse) Fail(f, t string, v ...interface{}) rinq.Failure {
	err := r.res.Fail(f, t, v...)
	r.logFailure(f, nil)
	return err
}

func (r *loggingResponse) Close() bool {
	if r.res.Close() {
		r.logSuccess(nil)
		return true
	}

	return false
}

func (r *loggingResponse) logSuccess(payload *rinq.Payload) {
	r.logger.Log(
		"%s handled %s '%s' command from %s successfully (%dms %d/i %d/o) [%s]",
		r.peerID.ShortString(),
		r.req.Namespace,
		r.req.Command,
		r.req.Source.Ref().ShortString(),
		time.Now().Sub(r.startedAt)/time.Millisecond,
		r.req.Payload.Len(),
		payload.Len(),
		r.traceID,
	)
}
func (r *loggingResponse) logFailure(failureType string, payload *rinq.Payload) {
	r.logger.Log(
		"%s handled %s '%s' command from %s: '%s' failure (%dms %d/i %d/o) [%s]",
		r.peerID.ShortString(),
		r.req.Namespace,
		r.req.Command,
		r.req.Source.Ref().ShortString(),
		failureType,
		time.Now().Sub(r.startedAt)/time.Millisecond,
		r.req.Payload.Len(),
		payload.Len(),
		r.traceID,
	)
}

func (r *loggingResponse) logError(err error) {
	r.logger.Log(
		"%s handled %s '%s' command from %s: '%s' error (%dms %d/i 0/o) [%s]",
		r.peerID.ShortString(),
		r.req.Namespace,
		r.req.Command,
		r.req.Source.Ref().ShortString(),
		err,
		time.Now().Sub(r.startedAt)/time.Millisecond,
		r.req.Payload.Len(),
		r.traceID,
	)
}
