package amqputil

import (
	"bytes"
	"context"
	"errors"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/rinq/internal/bufferpool"
	"github.com/streadway/amqp"
)

const (
	// spanContextHeader contains the serialied OpenTracing span context.
	spanContextHeader = "sc"
)

// PackSpanContext packs a serialized "span context" into the headers of msg
// based on the span in ctx, if any.
func PackSpanContext(ctx context.Context, msg *amqp.Publishing) error {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return nil
	}

	buf := bufferpool.Get()

	if err := span.Tracer().Inject(
		span.Context(),
		opentracing.Binary,
		buf,
	); err != nil {
		return err
	}

	if buf.Len() > 0 {
		if msg.Headers == nil {
			msg.Headers = amqp.Table{}
		}

		msg.Headers[spanContextHeader] = buf.Bytes()
	}

	return nil
}

// UnpackSpanContext extracts a span context from the headers of msg. If no
// span context is packed in the headers, nil is returned.
func UnpackSpanContext(msg *amqp.Delivery, t opentracing.Tracer) (opentracing.SpanContext, error) {
	v, ok := msg.Headers[spanContextHeader]
	if !ok {
		return nil, nil
	}

	b, ok := v.([]byte)
	if !ok {
		return nil, errors.New("span context header is not a byte slice")
	}

	buf := bytes.NewBuffer(b)
	defer bufferpool.Put(buf)

	sc, err := t.Extract(opentracing.Binary, buf)

	if err == opentracing.ErrSpanContextNotFound {
		return nil, nil
	}

	return sc, err
}
