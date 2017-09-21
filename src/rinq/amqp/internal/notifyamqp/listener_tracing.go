package notifyamqp

import (
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/rinq/rinq-go/src/rinq/amqp/internal/amqputil"
	"github.com/streadway/amqp"
)

func traceUnpackSpanOptions(msg *amqp.Delivery, t opentracing.Tracer) (opts []opentracing.StartSpanOption, err error) {
	sc, err := amqputil.UnpackSpanContext(msg, t)

	if err == nil {
		opts = append(opts, ext.SpanKindConsumer)

		if sc != nil {
			opts = append(opts, opentracing.FollowsFrom(sc))
		}
	}

	return
}
