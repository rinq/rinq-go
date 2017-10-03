package rinq_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("Constraint", func() {
	Describe("And", func() {
		It("is equivalent to rinq.And", func() {
			a := rinq.Equal("a", "1")
			b := rinq.Equal("b", "1")

			Expect(a.And(b)).To(Equal(rinq.And(a, b)))
		})
	})

	Describe("Or", func() {
		It("is equivalent to rinq.Or", func() {
			a := rinq.Equal("a", "1")
			b := rinq.Equal("b", "1")

			Expect(a.Or(b)).To(Equal(rinq.Or(a, b)))
		})
	})

	Describe("String", func() {
		DescribeTable(
			"returns an appropriate string representation",
			func(con rinq.Constraint, s string) {
				Expect(con.String()).To(Equal(s))
			},

			Entry(
				"Within()",
				rinq.Within(
					"ns",
					rinq.Equal("a", "1"),
				),
				"ns::{a=1}",
			),
			Entry(
				"Within() with multiple values",
				rinq.Within(
					"ns",
					rinq.Equal("a", "1"),
					rinq.Equal("b", "2"),
				),
				"ns::{a=1, b=2}",
			),

			Entry(
				"Equal()",
				rinq.Equal("a", "1"),
				"{a=1}",
			),
			Entry(
				"NotEqual()",
				rinq.NotEqual("a", "1"),
				"{a!=1}",
			),

			Entry(
				"Empty()",
				rinq.Empty("a"),
				"{!a}",
			),
			Entry(
				"NotEmpty()",
				rinq.NotEmpty("a"),
				"{a}",
			),

			Entry(
				"Not()",
				rinq.Not(rinq.Equal("a", "1")),
				"{! a=1}",
			),
			Entry(
				"Not() with compound expression",
				rinq.Not(
					rinq.And(
						rinq.Equal("a", "1"),
						rinq.Equal("b", "2"),
					),
				),
				"{! {a=1, b=2}}",
			),

			Entry(
				"And()",
				rinq.And(
					rinq.Equal("a", "1"),
				),
				"{a=1}",
			),
			Entry(
				"And() with multiple values",
				rinq.And(
					rinq.Equal("a", "1"),
					rinq.Equal("b", "2"),
				),
				"{a=1, b=2}",
			),

			Entry(
				"Or()",
				rinq.Or(
					rinq.Equal("a", "1"),
				),
				"{a=1}",
			),
			Entry(
				"Or() with multiple values",
				rinq.Or(
					rinq.Equal("a", "1"),
					rinq.Equal("b", "2"),
				),
				"{a=1|b=2}",
			),

			Entry(
				"nested compound expression",
				rinq.And(
					rinq.Equal("a", "1"),
					rinq.Or(
						rinq.Equal("b", "2"),
						rinq.Equal("c", "3"),
					),
				),
				"{a=1, {b=2|c=3}}",
			),
		)
	})
})
