package traceutil

import (
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

var (
	invokerCallEvent      = log.String("event", "call")
	invokerCallAsyncEvent = log.String("event", "call-async")
	invokerExecuteEvent   = log.String("event", "execute")

	invokerErrorSourceClient = log.String("error.source", "client")
	invokerErrorSourceServer = log.String("error.source", "server")

	invokerSuccessEvent = log.String("event", "success")
	invokerFailureEvent = log.String("event", "failure")

	serverRequestEvent  = log.String("event", "request")
	serverResponseEvent = log.String("event", "response")
)

const (
	commandKey = "command"
	serverKey  = "server"
)

// SetupCommand configures span as a command-related span.
func SetupCommand(
	s opentracing.Span,
	id ident.MessageID,
	ns string,
	cmd string,
) {
	s.SetOperationName(ns + "::" + cmd)

	s.SetTag(subsystemKey, "command")
	s.SetTag(messageIDKey, id.String())
	s.SetTag(namespaceKey, ns)
	s.SetTag(commandKey, cmd)
}

// LogInvokerCall logs information about a "call" style invocation to s.
func LogInvokerCall(s opentracing.Span, p *rinq.Payload) {
	s.LogFields(
		invokerCallEvent,
		log.Int(payloadSizeKey, p.Len()),
	)
}

// LogInvokerCallAsync logs information about a "call-sync" style invocation to s.
func LogInvokerCallAsync(span opentracing.Span, p *rinq.Payload) {
	span.LogFields(
		invokerCallAsyncEvent,
		log.Int(payloadSizeKey, p.Len()),
	)
}

// LogInvokerExecute logs information about an "execute" style invoation to s.
func LogInvokerExecute(span opentracing.Span, p *rinq.Payload) {
	span.LogFields(
		invokerExecuteEvent,
		log.Int(payloadSizeKey, p.Len()),
	)
}

// LogInvokerSuccess logs information about a successful command response to s.
func LogInvokerSuccess(span opentracing.Span, p *rinq.Payload) {
	span.LogFields(
		invokerSuccessEvent,
		log.Int(payloadSizeKey, p.Len()),
	)
}

// LogInvokerError logs information about err to s.
func LogInvokerError(s opentracing.Span, err error) {
	ext.Error.Set(s, true)

	switch e := err.(type) {
	case rinq.Failure:
		s.LogFields(
			invokerFailureEvent,
			log.String("error.kind", e.Type),
			log.String("message", e.Message),
			invokerErrorSourceServer,
			log.Int(payloadSizeKey, e.Payload.Len()),
		)

	case rinq.CommandError:
		s.LogFields(
			errorEvent,
			log.String("message", e.Error()),
			invokerErrorSourceServer,
		)

	default:
		s.LogFields(
			errorEvent,
			log.String("message", e.Error()),
			invokerErrorSourceClient,
		)
	}
}

// LogServerRequest logs information about an incoming command request to s.
func LogServerRequest(s opentracing.Span, peerID ident.PeerID, p *rinq.Payload) {
	s.SetTag(serverKey, peerID.String())

	s.LogFields(
		serverRequestEvent,
		log.Int(payloadSizeKey, p.Len()),
	)
}

// LogServerSuccess logs information about a successful command response to s.
func LogServerSuccess(span opentracing.Span, p *rinq.Payload) {
	span.LogFields(
		serverResponseEvent,
		log.Int(payloadSizeKey, p.Len()),
	)
}

// LogServerError logs information about err to s.
func LogServerError(s opentracing.Span, err error) {
	switch e := err.(type) {
	case rinq.Failure:
		s.LogFields(
			serverResponseEvent,
			log.String("error.kind", e.Type),
			log.String("message", e.Message),
			log.Int(payloadSizeKey, e.Payload.Len()),
		)

	default:
		ext.Error.Set(s, true)

		s.LogFields(
			serverResponseEvent,
			log.String("message", e.Error()),
		)
	}
}
