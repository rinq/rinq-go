package overpassamqp

import (
	"time"

	"github.com/over-pass/overpass-go/src/overpass"
)

// responder wraps a parent responder to add logging
type responder struct {
	parent    overpass.Responder
	peerID    overpass.PeerID
	corrID    string
	request   overpass.Command
	logger    overpass.Logger
	startedAt time.Time
}

func newResponder(
	parent overpass.Responder,
	peerID overpass.PeerID,
	corrID string,
	request overpass.Command,
	logger overpass.Logger,
) overpass.Responder {
	return &responder{
		parent:    parent,
		peerID:    peerID,
		corrID:    corrID,
		request:   request,
		logger:    logger,
		startedAt: time.Now(),
	}
}

func (r *responder) IsRequired() bool {
	return r.parent.IsRequired()
}

func (r *responder) IsClosed() bool {
	return r.parent.IsClosed()
}

func (r *responder) Done(payload *overpass.Payload) {
	r.parent.Done(payload)
	r.logSuccess(payload)
}

func (r *responder) Error(err error) {
	r.parent.Error(err)

	if failure, ok := err.(overpass.Failure); ok {
		r.logFailure(failure.Type, failure.Payload)
	} else {
		r.logError(err)
	}
}

func (r *responder) Fail(failureType, message string) {
	r.parent.Fail(failureType, message)
	r.logFailure(failureType, nil)
}

func (r *responder) Close() {
	first := !r.IsClosed()
	r.parent.Close()

	if first {
		r.logSuccess(nil)
	}
}

func (r *responder) logSuccess(payload *overpass.Payload) {
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
func (r *responder) logFailure(failureType string, payload *overpass.Payload) {
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

func (r *responder) logError(err error) {
	r.logger.Log(
		"%s handled %s '%s' command from %s: '%s' error (%dms %d/i 0)/o [%s]",
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
