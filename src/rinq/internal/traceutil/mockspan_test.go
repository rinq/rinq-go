package traceutil_test

import (
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type mockSpan struct {
	opentracing.Span

	operationName string
	log           [][]log.Field
	tags          map[string]interface{}
}

// Sets or changes the operation name.
func (s *mockSpan) SetOperationName(operationName string) opentracing.Span {
	s.operationName = operationName
	return s
}

// Adds a tag to the span.
//
// If there is a pre-existing tag set for `key`, it is overwritten.
//
// Tag values can be numeric types, strings, or bools. The behavior of
// other tag value types is undefined at the OpenTracing level. If a
// tracing system does not know how to handle a particular value type, it
// may ignore the tag, but shall not panic.
func (s *mockSpan) SetTag(key string, value interface{}) opentracing.Span {
	if s.tags == nil {
		s.tags = map[string]interface{}{}
	}

	s.tags[key] = value

	return s
}

// LogFields is an efficient and type-checked way to record key:value
// logging data about a Span, though the programming interface is a little
// more verbose than LogKV(). Here's an example:
//
//    span.LogFields(
//        log.String("event", "soft error"),
//        log.String("type", "cache timeout"),
//        log.Int("waited.millis", 1500))
//
// Also see Span.FinishWithOptions() and FinishOptions.BulkLogData.
func (s *mockSpan) LogFields(fields ...log.Field) {
	s.log = append(s.log, fields)
}
