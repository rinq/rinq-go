package marshaling

import (
	"bytes"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/internal/x/bufferpool"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/refactor/amqpx"
	"github.com/streadway/amqp"
)

const (
	// spanContextHeader contains the serialied OpenTracing span context.
	spanContextHeader = "sc"
)

// PackSpanContext packs a serialized "span context" into the headers of msg
// based on the span in ctx, if any. buf is a buffer used for encoding which
// must be valid until msg is published.
func PackSpanContext(
	msg *amqp.Publishing,
	t opentracing.Tracer,
	sc opentracing.SpanContext,
	buf *bytes.Buffer,
) error {
	if err := t.Inject(
		sc,
		opentracing.Binary,
		buf,
	); err != nil {
		return err
	}

	if buf.Len() > 0 {
		amqpx.SetHeader(msg, spanContextHeader, buf.Bytes())
	}

	return nil
}

// UnpackSpanContext extracts a span context from the headers of msg. If no
// span context is packed in the headers, nil is returned.
func UnpackSpanContext(msg *amqp.Delivery, t opentracing.Tracer) (opentracing.SpanContext, error) {
	b, ok, err := amqpx.GetHeaderBytesOptional(msg, spanContextHeader)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, nil
	}

	buf := bytes.NewBuffer(b)
	defer bufferpool.Put(buf)

	sc, err := t.Extract(opentracing.Binary, buf)

	if err == opentracing.ErrSpanContextNotFound {
		return nil, nil
	}

	return sc, err
}
