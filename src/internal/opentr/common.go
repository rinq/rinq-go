package opentr

import (
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

var (
	successEvent = log.String("event", "success")
	errorEvent   = log.String("event", "error")
)

// AddTraceID configures span s to have traceID set to the given id.
func AddTraceID(s opentracing.Span, id string) {
	if id != "" {
		s.SetTag("traceID", id)
	}
}
