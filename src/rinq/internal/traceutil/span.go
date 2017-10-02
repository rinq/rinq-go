package traceutil

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/rinq/rinq-go/src/rinq"
)

// CommonSpanOptions contains span options that should be applied to any
// spans started by Rinq.
var CommonSpanOptions = []opentracing.StartSpanOption{
	opentracing.Tag{
		Key:   string(ext.Component),
		Value: "rinq-go/" + rinq.Version,
	},
}

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

	opts = append(opts, CommonSpanOptions...)

	span := tracer.StartSpan("", opts...)

	return span, opentracing.ContextWithSpan(ctx, span)
}
