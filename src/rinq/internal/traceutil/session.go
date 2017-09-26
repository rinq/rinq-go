package traceutil

import (
	"bytes"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
	"github.com/rinq/rinq-go/src/rinq/internal/bufferpool"
)

const (
	fetchOp   = "session fetch"
	updateOp  = "session update"
	destroyOp = "session destroy"
)

var (
	fetchEvent   = log.String("event", "fetch")
	updateEvent  = log.String("event", "update")
	destroyEvent = log.String("event", "destroy")
)

func setupSessionCommand(s opentracing.Span, op string, sessID ident.SessionID) {
	s.SetOperationName(op)

	s.SetTag("subsystem", "session")
	s.SetTag("session", sessID.String())
}

// SetupSessionFetch configures s as an attribute fetch operation.
func SetupSessionFetch(s opentracing.Span, sessID ident.SessionID) {
	setupSessionCommand(s, fetchOp, sessID)
}

// SetupSessionUpdate configures s as an attribute update operation.
func SetupSessionUpdate(s opentracing.Span, sessID ident.SessionID) {
	setupSessionCommand(s, updateOp, sessID)
}

// SetupSessionDestroy configures s as a destroy operation.
func SetupSessionDestroy(s opentracing.Span, sessID ident.SessionID) {
	setupSessionCommand(s, destroyOp, sessID)
}

// LogSessionFetchRequest logs information about a session fetch attempt to s.
func LogSessionFetchRequest(s opentracing.Span, keys []string) {
	fields := []log.Field{
		fetchEvent,
	}

	if len(keys) != 0 {
		fields = append(fields, lazyString("keys", func() string {
			return strings.Join(keys, ", ")
		}))
	}

	s.LogFields(fields...)
}

// LogSessionFetchSuccess logs information about a successful session fetch to s.
func LogSessionFetchSuccess(s opentracing.Span, rev ident.Revision, attrs []attrmeta.Attr) {
	fields := []log.Field{
		successEvent,
		log.Uint32("rev", uint32(rev)),
	}

	if len(attrs) != 0 {
		fields = append(fields, lazyString("attributes", func() string {
			buf := bufferpool.Get()
			defer bufferpool.Put(buf)
			attrmeta.WriteSlice(buf, attrs)
			return buf.String()
		}))
	}

	s.LogFields(fields...)
}

// LogSessionUpdateRequest logs information about a session update attempt to s.
func LogSessionUpdateRequest(s opentracing.Span, rev ident.Revision, attrs []rinq.Attr) {
	fields := []log.Field{
		updateEvent,
		log.Uint32("rev", uint32(rev)),
	}

	if len(attrs) != 0 {
		fields = append(fields, lazyString("changes", func() string {
			buf := bufferpool.Get()
			defer bufferpool.Put(buf)

			for _, attr := range attrs {
				if buf.Len() > 0 {
					buf.WriteString(", ")
				}

				buf.WriteString(attr.String())
			}

			return buf.String()
		}))
	}

	s.LogFields(fields...)
}

// LogSessionUpdateSuccess logs information about a successful session update to s.
func LogSessionUpdateSuccess(s opentracing.Span, rev ident.Revision, diff *bytes.Buffer) {
	fields := []log.Field{
		successEvent,
		log.Uint32("rev", uint32(rev)),
	}

	if diff.Len() != 0 {
		fields = append(
			fields,
			log.String("diff", diff.String()),
		)
	}

	s.LogFields(fields...)
}

// LogSessionUpdateRequest logs information about a session destroy attempt to s.
func LogSessionDestroyRequest(s opentracing.Span, rev ident.Revision) {
	s.LogFields(
		destroyEvent,
		log.Uint32("rev", uint32(rev)),
	)
}

// LogSessionUpdateSuccess logs information about a successful destroy attempt to s.
func LogSessionDestroySuccess(s opentracing.Span) {
	s.LogFields(
		successEvent,
	)
}

// LogSessionError logs information about an error during a session operation.
func LogSessionError(s opentracing.Span, err error) {
	switch e := err.(type) {
	case rinq.Failure:
		s.LogFields(
			log.String("event", e.Type),
		)

	default:
		ext.Error.Set(s, true)

		s.LogFields(
			errorEvent,
			log.String("message", e.Error()),
		)
	}
}
