package constraint_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq/constraint"
)

var _ = Describe("Constraint", func() {
	Describe("And", func() {
		It("is equivalent to constraint.And", func() {
			a := constraint.Equal("a", "1")
			b := constraint.Equal("b", "1")

			Expect(a.And(b)).To(Equal(constraint.And(a, b)))
		})
	})

	Describe("Or", func() {
		It("is equivalent to constraint.Or", func() {
			a := constraint.Equal("a", "1")
			b := constraint.Equal("b", "1")

			Expect(a.Or(b)).To(Equal(constraint.Or(a, b)))
		})
	})

	Describe("String", func() {
		DescribeTable(
			"returns an appropriate string representation",
			func(con constraint.Constraint, s string) {
				Expect(con.String()).To(Equal(s))
			},

			Entry(
				"Within()",
				constraint.Within(
					"ns",
					constraint.Equal("a", "1"),
				),
				"ns::{a=1}",
			),
			Entry(
				"Within() with multiple values",
				constraint.Within(
					"ns",
					constraint.Equal("a", "1"),
					constraint.Equal("b", "2"),
				),
				"ns::{a=1, b=2}",
			),

			Entry(
				"Equal()",
				constraint.Equal("a", "1"),
				"{a=1}",
			),
			Entry(
				"NotEqual()",
				constraint.NotEqual("a", "1"),
				"{a!=1}",
			),

			Entry(
				"Empty()",
				constraint.Empty("a"),
				"{!a}",
			),
			Entry(
				"NotEmpty()",
				constraint.NotEmpty("a"),
				"{a}",
			),

			Entry(
				"Not()",
				constraint.Not(constraint.Equal("a", "1")),
				"{! a=1}",
			),
			Entry(
				"Not() with compound expression",
				constraint.Not(
					constraint.And(
						constraint.Equal("a", "1"),
						constraint.Equal("b", "2"),
					),
				),
				"{! {a=1, b=2}}",
			),

			Entry(
				"And()",
				constraint.And(
					constraint.Equal("a", "1"),
				),
				"{a=1}",
			),
			Entry(
				"And() with multiple values",
				constraint.And(
					constraint.Equal("a", "1"),
					constraint.Equal("b", "2"),
				),
				"{a=1, b=2}",
			),

			Entry(
				"Or()",
				constraint.Or(
					constraint.Equal("a", "1"),
				),
				"{a=1}",
			),
			Entry(
				"Or() with multiple values",
				constraint.Or(
					constraint.Equal("a", "1"),
					constraint.Equal("b", "2"),
				),
				"{a=1|b=2}",
			),

			Entry(
				"nested compound expression",
				constraint.And(
					constraint.Equal("a", "1"),
					constraint.Or(
						constraint.Equal("b", "2"),
						constraint.Equal("c", "3"),
					),
				),
				"{a=1, {b=2|c=3}}",
			),
		)
	})
})
