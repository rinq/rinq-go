package traceutil_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/opentracing/opentracing-go/log"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	. "github.com/rinq/rinq-go/src/rinq/internal/traceutil"
)

var _ = Describe("SetupCommand", func() {
	It("sets the operation name", func() {
		span := &mockSpan{}

		SetupCommand(span, ident.MessageID{}, "<ns>", "<cmd>")

		Expect(span.operationName).To(Equal("<ns>::<cmd>"))
	})

	It("sets the appropriate tags", func() {
		span := &mockSpan{}

		msgID := ident.NewPeerID().Session(1).At(0).Message(27)

		SetupCommand(span, msgID, "<ns>", "<cmd>")

		Expect(span.tags).To(Equal(map[string]interface{}{
			"subsystem":  "command",
			"message_id": msgID.String(),
			"namespace":  "<ns>",
			"command":    "<cmd>",
		}))
	})
})

var _ = Describe("LogInvokerCall", func() {
	It("logs the appropriate fields", func() {
		span := &mockSpan{}

		p := rinq.NewPayloadFromBytes(make([]byte, 4))
		defer p.Close()

		LogInvokerCall(span, p)

		Expect(span.log).To(Equal(
			[][]log.Field{
				{
					log.String("event", "call"),
					log.Int("payload.size", 4),
				},
			},
		))
	})
})

var _ = Describe("LogInvokerCallAsync", func() {
	It("logs the appropriate fields", func() {
		span := &mockSpan{}

		p := rinq.NewPayloadFromBytes(make([]byte, 4))
		defer p.Close()

		LogInvokerCallAsync(span, p)

		Expect(span.log).To(Equal(
			[][]log.Field{
				{
					log.String("event", "call-async"),
					log.Int("payload.size", 4),
				},
			},
		))
	})
})

var _ = Describe("LogInvokerExecute", func() {
	It("logs the appropriate fields", func() {
		span := &mockSpan{}

		p := rinq.NewPayloadFromBytes(make([]byte, 4))
		defer p.Close()

		LogInvokerExecute(span, p)

		Expect(span.log).To(Equal(
			[][]log.Field{
				{
					log.String("event", "execute"),
					log.Int("payload.size", 4),
				},
			},
		))
	})
})

var _ = Describe("LogInvokerSuccess", func() {
	It("logs the appropriate fields", func() {
		span := &mockSpan{}

		p := rinq.NewPayloadFromBytes(make([]byte, 4))
		defer p.Close()

		LogInvokerSuccess(span, p)

		Expect(span.log).To(Equal(
			[][]log.Field{
				{
					log.String("event", "success"),
					log.Int("payload.size", 4),
				},
			},
		))
	})
})

var _ = Describe("LogInvokerError", func() {
	Context("when the error is a failure", func() {
		err := rinq.Failure{
			Type:    "<type>",
			Message: "<message>",
			Payload: rinq.NewPayloadFromBytes(make([]byte, 4)),
		}

		It("logs the appropriate fields", func() {
			span := &mockSpan{}
			LogInvokerError(span, err)

			Expect(span.log).To(Equal(
				[][]log.Field{
					{
						log.String("event", "failure"),
						log.String("error.kind", "<type>"),
						log.String("message", "<message>"),
						log.String("error.source", "server"),
						log.Int("payload.size", 4),
					},
				},
			))
		})

		It("sets the error tag", func() {
			span := &mockSpan{}
			LogInvokerError(span, err)

			Expect(span.tags["error"]).To(BeTrue())
		})
	})

	Context("when the error is a server-side error", func() {
		err := rinq.CommandError("<error>")

		It("logs the appropriate fields", func() {
			span := &mockSpan{}
			LogInvokerError(span, err)

			Expect(span.log).To(Equal(
				[][]log.Field{
					{
						log.String("event", "error"),
						log.String("message", "<error>"),
						log.String("error.source", "server"),
					},
				},
			))
		})

		It("sets the error tag", func() {
			span := &mockSpan{}
			LogInvokerError(span, err)

			Expect(span.tags["error"]).To(BeTrue())
		})
	})

	Context("when the error is a client-side error", func() {
		err := errors.New("<error>")

		It("logs the appropriate fields", func() {
			span := &mockSpan{}
			LogInvokerError(span, err)

			Expect(span.log).To(Equal(
				[][]log.Field{
					{
						log.String("event", "error"),
						log.String("message", "<error>"),
						log.String("error.source", "client"),
					},
				},
			))
		})

		It("sets the error tag", func() {
			span := &mockSpan{}
			LogInvokerError(span, err)

			Expect(span.tags["error"]).To(BeTrue())
		})
	})
})

var _ = Describe("LogServerRequest", func() {
	It("sets the server tag", func() {
		span := &mockSpan{}

		peerID := ident.NewPeerID()

		p := rinq.NewPayloadFromBytes(make([]byte, 4))
		defer p.Close()

		LogServerRequest(span, peerID, p)

		Expect(span.tags["server"]).To(Equal(peerID.String()))
	})

	It("logs the appropriate fields", func() {
		span := &mockSpan{}

		peerID := ident.NewPeerID()

		p := rinq.NewPayloadFromBytes(make([]byte, 4))
		defer p.Close()

		LogServerRequest(span, peerID, p)

		Expect(span.log).To(Equal(
			[][]log.Field{
				{
					log.String("event", "request"),
					log.Int("payload.size", 4),
				},
			},
		))
	})
})

var _ = Describe("LogServerSuccess", func() {
	It("logs the appropriate fields", func() {
		span := &mockSpan{}

		p := rinq.NewPayloadFromBytes(make([]byte, 4))
		defer p.Close()

		LogServerSuccess(span, p)

		Expect(span.log).To(Equal(
			[][]log.Field{
				{
					log.String("event", "response"),
					log.Int("payload.size", 4),
				},
			},
		))
	})
})

var _ = Describe("LogServerError", func() {
	Context("when the error is a failure", func() {
		err := rinq.Failure{
			Type:    "<type>",
			Message: "<message>",
			Payload: rinq.NewPayloadFromBytes(make([]byte, 4)),
		}

		It("logs the appropriate fields", func() {
			span := &mockSpan{}
			LogServerError(span, err)

			Expect(span.log).To(Equal(
				[][]log.Field{
					{
						log.String("event", "response"),
						log.String("error.kind", "<type>"),
						log.String("message", "<message>"),
						log.Int("payload.size", 4),
					},
				},
			))
		})

		It("does not set the error tag", func() {
			span := &mockSpan{}
			LogServerError(span, err)

			Expect(span.tags["error"]).To(BeNil())
		})
	})

	Context("when the error is not a failure", func() {
		err := errors.New("<error>")

		It("logs the appropriate fields", func() {
			span := &mockSpan{}
			LogServerError(span, err)

			Expect(span.log).To(Equal(
				[][]log.Field{
					{
						log.String("event", "response"),
						log.String("message", "<error>"),
					},
				},
			))
		})

		It("sets the error tag", func() {
			span := &mockSpan{}
			LogServerError(span, err)

			Expect(span.tags["error"]).To(BeTrue())
		})
	})
})
