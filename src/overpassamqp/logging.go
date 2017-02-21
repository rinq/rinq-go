package overpassamqp

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
		"%s stopped listening for command requests in '%s' namespace",
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

type loggingResponder struct {
	parent    overpass.Responder
	peerID    overpass.PeerID
	corrID    string
	request   overpass.Command
	logger    overpass.Logger
	startedAt time.Time
}

func newLoggingResponder(
	parent overpass.Responder,
	peerID overpass.PeerID,
	corrID string,
	request overpass.Command,
	logger overpass.Logger,
) overpass.Responder {
	return &loggingResponder{
		parent:    parent,
		peerID:    peerID,
		corrID:    corrID,
		request:   request,
		logger:    logger,
		startedAt: time.Now(),
	}
}

func (r *loggingResponder) IsRequired() bool {
	return r.parent.IsRequired()
}

func (r *loggingResponder) IsClosed() bool {
	return r.parent.IsClosed()
}

func (r *loggingResponder) Done(payload *overpass.Payload) {
	r.parent.Done(payload)
	r.logSuccess(payload)
}

func (r *loggingResponder) Error(err error) {
	r.parent.Error(err)

	if failure, ok := err.(overpass.Failure); ok {
		r.logFailure(failure.Type, failure.Payload)
	} else {
		r.logError(err)
	}
}

func (r *loggingResponder) Fail(failureType, message string) {
	r.parent.Fail(failureType, message)
	r.logFailure(failureType, nil)
}

func (r *loggingResponder) Close() {
	first := !r.IsClosed()
	r.parent.Close()

	if first {
		r.logSuccess(nil)
	}
}

func (r *loggingResponder) logSuccess(payload *overpass.Payload) {
	r.logger.Log(
		"%s handled %s '%s' command from %s successfully (%dms %d/i %d/o) [%s]",
		r.peerID.ShortString(),
		r.request.Namespace,
		r.request.Command,
		r.request.Source.Ref().ShortString(),
		time.Now().Sub(r.startedAt)/time.Millisecond,
		r.request.Payload.Len(),
		payload.Len(),
		r.corrID,
	)
}
func (r *loggingResponder) logFailure(failureType string, payload *overpass.Payload) {
	r.logger.Log(
		"%s handled %s '%s' command from %s: '%s' failure (%dms %d/i %d/o) [%s]",
		r.peerID.ShortString(),
		r.request.Namespace,
		r.request.Command,
		r.request.Source.Ref().ShortString(),
		failureType,
		time.Now().Sub(r.startedAt)/time.Millisecond,
		r.request.Payload.Len(),
		payload.Len(),
		r.corrID,
	)
}

func (r *loggingResponder) logError(err error) {
	r.logger.Log(
		"%s handled %s '%s' command from %s: '%s' error (%dms %d/i 0/o) [%s]",
		r.peerID.ShortString(),
		r.request.Namespace,
		r.request.Command,
		r.request.Source.Ref().ShortString(),
		err,
		time.Now().Sub(r.startedAt)/time.Millisecond,
		r.request.Payload.Len(),
		r.corrID,
	)
}
