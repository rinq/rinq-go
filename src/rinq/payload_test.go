package rinq_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("Payload", func() {
	Describe("Clone", func() {
		DescribeTable(
			"returns a nil pointer",
			func(p *rinq.Payload) {
				defer p.Close()

				c := p.Clone()
				defer c.Close()

				Expect(c).To(BeNil())
			},
			Entry("nil pointer", nil),
			Entry("default value", &rinq.Payload{}),
			Entry("created from empty bytes", rinq.NewPayloadFromBytes(nil)),
			Entry("created from nil value", rinq.NewPayload(nil)),
		)

		DescribeTable(
			"returs an equivalent payload",
			func(p *rinq.Payload) {
				defer p.Close()

				c := p.Clone()
				defer c.Close()

				Expect(c.Value()).To(BeEquivalentTo(c.Value()))
			},
			Entry("created from bytes", rinq.NewPayloadFromBytes([]byte{24, 123})),
			Entry("created from value", rinq.NewPayload(123)),
		)
	})

	Describe("Bytes", func() {
		DescribeTable(
			"returns the expected binary representation",
			func(p *rinq.Payload, expected interface{}) {
				defer p.Close()

				buf := p.Bytes()

				if expected == nil {
					Expect(buf).To(BeNil())
				} else {
					Expect(buf).To(Equal(expected))
				}
			},
			Entry("nil pointer", nil, nil),
			Entry("default value", &rinq.Payload{}, nil),
			Entry("created from empty bytes", rinq.NewPayloadFromBytes(nil), nil),
			Entry("created from bytes", rinq.NewPayloadFromBytes([]byte{24, 123}), []byte{24, 123}),
			Entry("created from value", rinq.NewPayload(123), []byte{24, 123}),
		)
	})

	Describe("Len", func() {
		DescribeTable(
			"returns the binary byte length",
			func(p *rinq.Payload, expected int) {
				defer p.Close()

				Expect(p.Len()).To(Equal(expected))
			},
			Entry("nil pointer", nil, 0),
			Entry("default value", &rinq.Payload{}, 0),
			Entry("created from empty bytes", rinq.NewPayloadFromBytes(nil), 0),
			Entry("created from bytes", rinq.NewPayloadFromBytes([]byte{24, 123}), 2),
			Entry("created from value", rinq.NewPayload(123), 2),
		)
	})

	Describe("Value", func() {
		DescribeTable(
			"returns the expected value",
			func(p *rinq.Payload, expected interface{}) {
				defer p.Close()

				v := p.Value()

				if expected == nil {
					Expect(v).To(BeNil())
				} else {
					Expect(v).To(BeEquivalentTo(expected))
				}
			},
			Entry("nil pointer", nil, nil),
			Entry("default value", &rinq.Payload{}, nil),
			Entry("created from empty bytes", rinq.NewPayloadFromBytes(nil), nil),
			Entry("created from bytes", rinq.NewPayloadFromBytes([]byte{24, 123}), 123),
			Entry("created from value", rinq.NewPayload(123), 123),
		)
	})

	Describe("Decode", func() {
		DescribeTable(
			"returns the expected value",
			func(p *rinq.Payload, expected interface{}) {
				defer p.Close()

				var v interface{}
				err := p.Decode(&v)

				Expect(err).ShouldNot(HaveOccurred())

				if expected == nil {
					Expect(v).To(BeNil())
				} else {
					Expect(v).To(BeEquivalentTo(expected))
				}
			},
			Entry("nil pointer", nil, nil),
			Entry("default value", &rinq.Payload{}, nil),
			Entry("created from empty bytes", rinq.NewPayloadFromBytes(nil), nil),
			Entry("created from bytes", rinq.NewPayloadFromBytes([]byte{24, 123}), 123),
			Entry("created from value", rinq.NewPayload(123), 123),
		)

		It("can be called after Value() when created from bytes [regression]", func() {
			payload := rinq.NewPayloadFromBytes([]byte{103, 60, 118, 97, 108, 117, 101, 62})
			defer payload.Close()

			Expect(payload.Value()).To(Equal("<value>"))

			var value string
			err := payload.Decode(&value)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(value).To(Equal("<value>"))
		})
	})

	Describe("Close", func() {
		It("resets the payload value to nil", func() {
			p := rinq.NewPayload(123)
			p.Close()

			Expect(p.Value()).To(BeNil())
		})
	})

	Describe("Close", func() {
		It("returns a JSON representation of the payload", func() {
			p := rinq.NewPayload(map[interface{}]interface{}{"foo": "bar"})
			defer p.Close()

			Expect(p.String()).To(Equal(`{"foo":"bar"}`))
		})
	})
})

var _ = Describe("NewPayload", func() {
	DescribeTable(
		"returns nil when the value is nil",
		func(v interface{}) {
			p := rinq.NewPayload(v)
			defer p.Close()

			Expect(p).To(BeNil())
		},
		Entry("nil", nil),
		Entry("nil pointer", (*int)(nil)),
		Entry("nil interface", (interface{})(nil)),
	)

	DescribeTable(
		"returns a non-nil payload when the value is a nil container",
		func(v interface{}) {
			p := rinq.NewPayload(v)
			defer p.Close()

			Expect(p).NotTo(BeNil())
		},
		Entry("nil map", (map[int]int)(nil)),
		Entry("nil slice", ([]int)(nil)),
	)
})
