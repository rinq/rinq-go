package overpass_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/over-pass/overpass-go/src/overpass"
)

var _ = Describe("Payload", func() {
	Describe("NewPayload", func() {
		DescribeTable(
			"returns nil when the value is nil",
			func(v interface{}) {
				p := overpass.NewPayload(v)
				defer p.Close()

				Expect(p).To(BeNil())
			},
			Entry("nil", nil),
			Entry("nil channel", (*int)(nil)),
			Entry("nil function", (func())(nil)),
			Entry("nil map", (map[int]int)(nil)),
			Entry("nil pointer", (*int)(nil)),
			Entry("nil interface", (interface{})(nil)),
			Entry("nil slice", ([]int)(nil)),
		)
	})

	Describe("Clone", func() {
		DescribeTable(
			"returns a nil pointer",
			func(p *overpass.Payload) {
				defer p.Close()

				c := p.Clone()
				defer c.Close()

				Expect(c).To(BeNil())
			},
			Entry("nil pointer", nil),
			Entry("default value", &overpass.Payload{}),
			Entry("created from empty bytes", overpass.NewPayloadFromBytes(nil)),
			Entry("created from nil value", overpass.NewPayload(nil)),
		)

		DescribeTable(
			"returs an equivalent payload",
			func(p *overpass.Payload) {
				defer p.Close()

				c := p.Clone()
				defer c.Close()

				Expect(c.Value()).To(BeEquivalentTo(c.Value()))
			},
			Entry("created from bytes", overpass.NewPayloadFromBytes([]byte{24, 123})),
			Entry("created from value", overpass.NewPayload(123)),
		)
	})

	Describe("Bytes", func() {
		DescribeTable(
			"returns the expected binary representation",
			func(p *overpass.Payload, expected interface{}) {
				defer p.Close()

				buf := p.Bytes()

				if expected == nil {
					Expect(buf).To(BeNil())
				} else {
					Expect(buf).To(Equal(expected))
				}
			},
			Entry("nil pointer", nil, nil),
			Entry("default value", &overpass.Payload{}, nil),
			Entry("created from empty bytes", overpass.NewPayloadFromBytes(nil), nil),
			Entry("created from bytes", overpass.NewPayloadFromBytes([]byte{24, 123}), []byte{24, 123}),
			Entry("created from value", overpass.NewPayload(123), []byte{24, 123}),
		)
	})

	Describe("Len", func() {
		DescribeTable(
			"returns the binary byte length",
			func(p *overpass.Payload, expected int) {
				defer p.Close()

				Expect(p.Len()).To(Equal(expected))
			},
			Entry("nil pointer", nil, 0),
			Entry("default value", &overpass.Payload{}, 0),
			Entry("created from empty bytes", overpass.NewPayloadFromBytes(nil), 0),
			Entry("created from bytes", overpass.NewPayloadFromBytes([]byte{24, 123}), 2),
			Entry("created from value", overpass.NewPayload(123), 2),
		)
	})

	Describe("Value", func() {
		DescribeTable(
			"returns the expected value",
			func(p *overpass.Payload, expected interface{}) {
				defer p.Close()

				v := p.Value()

				if expected == nil {
					Expect(v).To(BeNil())
				} else {
					Expect(v).To(BeEquivalentTo(expected))
				}
			},
			Entry("nil pointer", nil, nil),
			Entry("default value", &overpass.Payload{}, nil),
			Entry("created from empty bytes", overpass.NewPayloadFromBytes(nil), nil),
			Entry("created from bytes", overpass.NewPayloadFromBytes([]byte{24, 123}), 123),
			Entry("created from value", overpass.NewPayload(123), 123),
		)
	})

	Describe("Decode", func() {
		DescribeTable(
			"returns the expected value",
			func(p *overpass.Payload, expected interface{}) {
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
			Entry("default value", &overpass.Payload{}, nil),
			Entry("created from empty bytes", overpass.NewPayloadFromBytes(nil), nil),
			Entry("created from bytes", overpass.NewPayloadFromBytes([]byte{24, 123}), 123),
			Entry("created from value", overpass.NewPayload(123), 123),
		)

		It("can be called after Value() when created from bytes [regression]", func() {
			payload := overpass.NewPayloadFromBytes([]byte{103, 60, 118, 97, 108, 117, 101, 62})
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
			p := overpass.NewPayload(123)
			p.Close()

			Expect(p.Value()).To(BeNil())
		})
	})

	Describe("Close", func() {
		It("returns a JSON representation of the payload", func() {
			p := overpass.NewPayload(map[interface{}]interface{}{"foo": "bar"})
			defer p.Close()

			Expect(p.String()).To(Equal(`{"foo":"bar"}`))
		})
	})
})
