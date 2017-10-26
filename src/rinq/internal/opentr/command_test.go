package opentr_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/attributes"
	. "github.com/rinq/rinq-go/src/rinq/internal/opentr"
)

var _ = Describe("SetupCommand", func() {
	It("sets the operation name", func() {
		span := &mockSpan{}

		SetupCommand(span, ident.MessageID{}, "<ns>", "<cmd>")

		Expect(span.operationName).To(Equal("<ns>::<cmd> command"))
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

		attrs := attributes.Catalog{
			"ns": {
				"foo": attributes.VAttr{
					Attr: rinq.Freeze("foo", "bar"),
				},
			},
		}

		p := rinq.NewPayloadFromBytes(make([]byte, 4))
		defer p.Close()

		LogInvokerCall(span, attrs, p)

		Expect(span.log).To(Equal(
			[]map[string]interface{}{
				{
					"event":      "call",
					"attributes": "ns::{foo@bar}",
					"size":       4,
				},
			},
		))
	})
})

var _ = Describe("LogInvokerCallAsync", func() {
	It("logs the appropriate fields", func() {
		span := &mockSpan{}

		attrs := attributes.Catalog{
			"ns": {
				"foo": attributes.VAttr{
					Attr: rinq.Freeze("foo", "bar"),
				},
			},
		}

		p := rinq.NewPayloadFromBytes(make([]byte, 4))
		defer p.Close()

		LogInvokerCallAsync(span, attrs, p)

		Expect(span.log).To(Equal(
			[]map[string]interface{}{
				{
					"event":      "call-async",
					"attributes": "ns::{foo@bar}",
					"size":       4,
				},
			},
		))
	})
})

var _ = Describe("LogInvokerExecute", func() {
	It("logs the appropriate fields", func() {
		span := &mockSpan{}

		attrs := attributes.Catalog{
			"ns": {
				"foo": attributes.VAttr{
					Attr: rinq.Freeze("foo", "bar"),
				},
			},
		}

		p := rinq.NewPayloadFromBytes(make([]byte, 4))
		defer p.Close()

		LogInvokerExecute(span, attrs, p)

		Expect(span.log).To(Equal(
			[]map[string]interface{}{
				{
					"event":      "execute",
					"attributes": "ns::{foo@bar}",
					"size":       4,
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
			[]map[string]interface{}{
				{
					"event": "success",
					"size":  4,
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
				[]map[string]interface{}{
					{
						"event":        "failure",
						"error.kind":   "<type>",
						"message":      "<message>",
						"error.source": "server",
						"size":         4,
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
				[]map[string]interface{}{
					{
						"event":        "error",
						"message":      "<error>",
						"error.source": "server",
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
				[]map[string]interface{}{
					{
						"event":        "error",
						"message":      "<error>",
						"error.source": "client",
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
	It("logs the appropriate fields", func() {
		span := &mockSpan{}

		peerID := ident.NewPeerID()

		p := rinq.NewPayloadFromBytes(make([]byte, 4))
		defer p.Close()

		LogServerRequest(span, peerID, p)

		Expect(span.log).To(Equal(
			[]map[string]interface{}{
				{
					"event":  "request",
					"server": peerID.String(),
					"size":   4,
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
			[]map[string]interface{}{
				{
					"event": "response",
					"size":  4,
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
				[]map[string]interface{}{
					{
						"event":      "response",
						"error.kind": "<type>",
						"message":    "<message>",
						"size":       4,
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
				[]map[string]interface{}{
					{
						"event":   "response",
						"message": "<error>",
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
