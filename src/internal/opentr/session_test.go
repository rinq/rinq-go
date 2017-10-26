package opentr_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/internal/attributes"
	. "github.com/rinq/rinq-go/src/internal/opentr"
)

var _ = Describe("SetupSessionFetch", func() {
	It("sets the operation name", func() {
		span := &mockSpan{}

		SetupSessionFetch(span, "<ns>", ident.SessionID{})

		Expect(span.operationName).To(Equal("session fetch"))
	})

	It("sets the appropriate tags", func() {
		span := &mockSpan{}

		sessID := ident.NewPeerID().Session(1)

		SetupSessionFetch(span, "<ns>", sessID)

		Expect(span.tags).To(Equal(map[string]interface{}{
			"subsystem": "session",
			"session":   sessID.String(),
			"namespace": "<ns>",
		}))
	})
})

var _ = Describe("LogSessionFetchRequest", func() {
	It("logs the appropriate fields", func() {
		span := &mockSpan{}

		LogSessionFetchRequest(span, []string{"a", "b"})

		Expect(span.log).To(Equal(
			[]map[string]interface{}{
				{
					"event": "fetch",
					"keys":  "{a, b}",
				},
			},
		))
	})
})

var _ = Describe("LogSessionFetchSuccess", func() {
	It("logs the appropriate fields", func() {
		span := &mockSpan{}

		attrs := attributes.VList{
			{Attr: rinq.Set("a", "1")},
			{Attr: rinq.Set("b", "2")},
		}

		LogSessionFetchSuccess(span, 23, attrs)

		Expect(span.log).To(Equal(
			[]map[string]interface{}{
				{
					"event":      "success",
					"rev":        uint32(23),
					"attributes": "{a=1, b=2}",
				},
			},
		))
	})
})

var _ = Describe("SetupSessionUpdate", func() {
	It("sets the operation name", func() {
		span := &mockSpan{}

		SetupSessionUpdate(span, "<ns>", ident.SessionID{})

		Expect(span.operationName).To(Equal("session update"))
	})

	It("sets the appropriate tags", func() {
		span := &mockSpan{}

		sessID := ident.NewPeerID().Session(1)

		SetupSessionUpdate(span, "<ns>", sessID)

		Expect(span.tags).To(Equal(map[string]interface{}{
			"subsystem": "session",
			"session":   sessID.String(),
			"namespace": "<ns>",
		}))
	})
})

var _ = Describe("LogSessionUpdateRequest", func() {
	It("logs the appropriate fields", func() {
		span := &mockSpan{}

		attrs := attributes.List{
			rinq.Set("a", "1"),
			rinq.Set("b", "2"),
		}

		LogSessionUpdateRequest(span, 23, attrs)

		Expect(span.log).To(Equal(
			[]map[string]interface{}{
				{
					"event":   "update",
					"rev":     uint32(23),
					"changes": "{a=1, b=2}",
				},
			},
		))
	})
})

var _ = Describe("LogSessionUpdateSuccess", func() {
	It("logs the appropriate fields", func() {
		span := &mockSpan{}

		diff := attributes.NewDiff("ns", 23)
		diff.Append(
			attributes.VAttr{Attr: rinq.Set("a", "1")},
			attributes.VAttr{Attr: rinq.Set("b", "2")},
		)

		LogSessionUpdateSuccess(span, 23, diff)

		Expect(span.log).To(Equal(
			[]map[string]interface{}{
				{
					"event": "success",
					"rev":   uint32(23),
					"diff":  diff.StringWithoutNamespace(),
				},
			},
		))
	})
})

var _ = Describe("SetupSessionClear", func() {
	It("sets the operation name", func() {
		span := &mockSpan{}

		SetupSessionClear(span, "<ns>", ident.SessionID{})

		Expect(span.operationName).To(Equal("session clear"))
	})

	It("sets the appropriate tags", func() {
		span := &mockSpan{}

		sessID := ident.NewPeerID().Session(1)

		SetupSessionClear(span, "<ns>", sessID)

		Expect(span.tags).To(Equal(map[string]interface{}{
			"subsystem": "session",
			"session":   sessID.String(),
			"namespace": "<ns>",
		}))
	})
})

var _ = Describe("LogSessionClearRequest", func() {
	It("logs the appropriate fields", func() {
		span := &mockSpan{}

		LogSessionClearRequest(span, 23)

		Expect(span.log).To(Equal(
			[]map[string]interface{}{
				{
					"event": "clear",
					"rev":   uint32(23),
				},
			},
		))
	})
})

var _ = Describe("LogSessionClearSuccess", func() {
	It("logs the appropriate fields", func() {
		span := &mockSpan{}

		diff := attributes.NewDiff("ns", 32)
		diff.Append(
			attributes.VAttr{Attr: rinq.Set("a", ""), UpdatedAt: 23},
			attributes.VAttr{Attr: rinq.Set("b", ""), UpdatedAt: 23},
		)

		LogSessionClearSuccess(span, 23, diff)

		Expect(span.log).To(Equal(
			[]map[string]interface{}{
				{
					"event": "success",
					"rev":   uint32(23),
					"diff":  diff.StringWithoutNamespace(),
				},
			},
		))
	})

	It("allows a nil diff", func() {
		span := &mockSpan{}

		LogSessionClearSuccess(span, 23, nil)

		Expect(span.log).To(Equal(
			[]map[string]interface{}{
				{
					"event": "success",
					"rev":   uint32(23),
				},
			},
		))
	})
})

var _ = Describe("SetupSessionDestroy", func() {
	It("sets the operation name", func() {
		span := &mockSpan{}

		SetupSessionDestroy(span, ident.SessionID{})

		Expect(span.operationName).To(Equal("session destroy"))
	})

	It("sets the appropriate tags", func() {
		span := &mockSpan{}

		sessID := ident.NewPeerID().Session(1)

		SetupSessionDestroy(span, sessID)

		Expect(span.tags).To(Equal(map[string]interface{}{
			"subsystem": "session",
			"session":   sessID.String(),
		}))
	})
})

var _ = Describe("LogSessionDestroyRequest", func() {
	It("logs the appropriate fields", func() {
		span := &mockSpan{}

		LogSessionDestroyRequest(span, 23)

		Expect(span.log).To(Equal(
			[]map[string]interface{}{
				{
					"event": "destroy",
					"rev":   uint32(23),
				},
			},
		))
	})
})

var _ = Describe("LogSessionDestroySuccess", func() {
	It("logs the appropriate fields", func() {
		span := &mockSpan{}

		LogSessionDestroySuccess(span)

		Expect(span.log).To(Equal(
			[]map[string]interface{}{
				{
					"event": "success",
				},
			},
		))
	})
})

var _ = Describe("LogSessionError", func() {
	Context("when the error is a failure", func() {
		err := rinq.Failure{
			Type:    "<type>",
			Message: "<message>",
			Payload: rinq.NewPayloadFromBytes(make([]byte, 4)),
		}

		It("logs the appropriate fields", func() {
			span := &mockSpan{}
			LogSessionError(span, err)

			Expect(span.log).To(Equal(
				[]map[string]interface{}{
					{
						"event": "<type>",
					},
				},
			))
		})

		It("does not set the error tag", func() {
			span := &mockSpan{}
			LogSessionError(span, err)

			Expect(span.tags["error"]).To(BeNil())
		})
	})

	Context("when the error is not a failure", func() {
		err := errors.New("<error>")

		It("logs the appropriate fields", func() {
			span := &mockSpan{}
			LogSessionError(span, err)

			Expect(span.log).To(Equal(
				[]map[string]interface{}{
					{
						"event":   "error",
						"message": "<error>",
					},
				},
			))
		})

		It("sets the error tag", func() {
			span := &mockSpan{}
			LogSessionError(span, err)

			Expect(span.tags["error"]).To(BeTrue())
		})
	})
})
