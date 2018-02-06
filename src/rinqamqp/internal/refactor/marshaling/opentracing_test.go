package marshaling_test

import (
	"bytes"
	"io"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/refactor/amqptest"
	. "github.com/rinq/rinq-go/src/rinqamqp/internal/refactor/marshaling"
	"github.com/streadway/amqp"
)

var _ = Describe("PackSpanContext and UnpackSpanContext", func() {
	It("transmits the span context", func() {
		t := tracer{}
		pub := &amqp.Publishing{}
		buf := &bytes.Buffer{}
		err := PackSpanContext(pub, t, spanContext("<span context>"), buf)
		Expect(err).ShouldNot(HaveOccurred())

		del := amqptest.PublishingToDelivery(pub)
		sc, err := UnpackSpanContext(del, t)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(sc).To(Equal(spanContext("<span context>")))
	})
})

type spanContext string

func (spanContext) ForeachBaggageItem(handler func(k, v string) bool) {
	panic("not implemented")
}

type tracer struct{}

func (tracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	panic("not implemented")
}

func (tracer) Inject(sm opentracing.SpanContext, format interface{}, carrier interface{}) error {
	if format != opentracing.Binary {
		panic("format must be binary")
	}

	_, err := io.WriteString(
		carrier.(io.Writer),
		string(sm.(spanContext)),
	)

	return err
}

func (tracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	if format != opentracing.Binary {
		panic("format must be binary")
	}

	buf, err := ioutil.ReadAll(carrier.(io.Reader))
	if err != nil {
		return nil, err
	}

	return spanContext(buf), nil
}
