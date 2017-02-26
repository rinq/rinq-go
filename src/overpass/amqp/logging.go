package amqp

import (
	"time"

	"github.com/over-pass/overpass-go/src/overpass"
)

func logStartedListening(
	logger overpass.Logger,
	peerID overpass.PeerID,
	namespace string,
) {
	logger.Log(
		"%s started listening for command requests in '%s' namespace",
		peerID.ShortString(),
		namespace,
	)
}

func logStoppedListening(
	logger overpass.Logger,
	peerID overpass.PeerID,
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
	req overpass.Request
	res overpass.Response

	peerID    overpass.PeerID
	traceID   string
	logger    overpass.Logger
	startedAt time.Time
}

func newLoggingResponse(
	req overpass.Request,
	res overpass.Response,
	peerID overpass.PeerID,
	traceID string,
	logger overpass.Logger,
) overpass.Response {
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

func (r *loggingResponse) Done(payload *overpass.Payload) {
	r.res.Done(payload)
	r.logSuccess(payload)
}

func (r *loggingResponse) Error(err error) {
	r.res.Error(err)

	if failure, ok := err.(overpass.Failure); ok {
		r.logFailure(failure.Type, failure.Payload)
	} else {
		r.logError(err)
	}
}

func (r *loggingResponse) Fail(failureType, message string) overpass.Failure {
	err := r.res.Fail(failureType, message)
	r.logFailure(failureType, nil)
	return err
}

func (r *loggingResponse) Close() bool {
	if r.res.Close() {
		r.logSuccess(nil)
		return true
	}

	return false
}

func (r *loggingResponse) logSuccess(payload *overpass.Payload) {
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
func (r *loggingResponse) logFailure(failureType string, payload *overpass.Payload) {
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
