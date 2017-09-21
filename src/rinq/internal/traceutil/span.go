package traceutil

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/rinq/rinq-go/src/rinq"
)

// FollowsFrom creates a new span that with a follows-from relationship to the
// span in ctx, if any.
func FollowsFrom(
	ctx context.Context,
	tracer opentracing.Tracer,
	opts ...opentracing.StartSpanOption,
) (opentracing.Span, context.Context) {
	return childSpanFromContext(
		ctx,
		tracer,
		opentracing.FollowsFrom,
		opts...,
	)
}

// ChildOf creates a new span that with a child-of relationship to the span in,
// if any.
func ChildOf(
	ctx context.Context,
	tracer opentracing.Tracer,
	opts ...opentracing.StartSpanOption,
) (opentracing.Span, context.Context) {
	return childSpanFromContext(
		ctx,
		tracer,
		opentracing.FollowsFrom,
		opts...,
	)
}

func childSpanFromContext(
	ctx context.Context,
	tracer opentracing.Tracer,
	rel func(opentracing.SpanContext) opentracing.SpanReference,
	opts ...opentracing.StartSpanOption,
) (opentracing.Span, context.Context) {
	parent := opentracing.SpanFromContext(ctx)

	if parent != nil {
		opts = append(opts, rel(parent.Context()))
		tracer = parent.Tracer()
	}

	span := tracer.StartSpan("", opts...)
	ext.Component.Set(span, "rinq/"+rinq.Version)

	return span, opentracing.ContextWithSpan(ctx, span)
}
