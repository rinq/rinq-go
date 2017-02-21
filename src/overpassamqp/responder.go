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

	r.logger.Log(
		"%s handled '%s' in '%s' namespace (%d bytes), done after %dms (%d bytes) [%s]",
		r.peerID.ShortString(),
		r.request.Command,
		r.request.Namespace,
		r.request.Payload.Len(),
		time.Now().Sub(r.startedAt)/time.Millisecond,
		payload.Len(),
		r.corrID,
	)
}

func (r *responder) Error(err error) {
	r.parent.Error(err)

	if failure, ok := err.(overpass.Failure); ok {
		r.logFailure(failure.Type, failure.Message)
	} else {
		r.logger.Log(
			"%s handled '%s' in '%s' namespace (%d bytes), errored after %dms (%s) [%s]",
			r.peerID.ShortString(),
			r.request.Command,
			r.request.Namespace,
			r.request.Payload.Len(),
			time.Now().Sub(r.startedAt)/time.Millisecond,
			err,
			r.corrID,
		)
	}
}

func (r *responder) Fail(failureType, message string) {
	r.parent.Fail(failureType, message)
	r.logFailure(failureType, message)
}

func (r *responder) Close() {
	first := !r.IsClosed()
	r.parent.Close()

	if first {
		r.logger.Log(
			"%s handled '%s' in '%s' namespace (%d bytes), closed after %dms [%s]",
			r.peerID.ShortString(),
			r.request.Command,
			r.request.Namespace,
			r.request.Payload.Len(),
			time.Now().Sub(r.startedAt)/time.Millisecond,
			r.corrID,
		)
	}
}

func (r *responder) logFailure(failureType, message string) {
	r.logger.Log(
		"%s handled '%s' in '%s' namespace (%d bytes), failed after %dms (%s: %s) [%s]",
		r.peerID.ShortString(),
		r.request.Command,
		r.request.Namespace,
		r.request.Payload.Len(),
		time.Now().Sub(r.startedAt)/time.Millisecond,
		failureType,
		message,
		r.corrID,
	)
}
