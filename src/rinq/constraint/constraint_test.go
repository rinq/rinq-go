package constraint_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq/constraint"
)

var _ = Describe("Constraint", func() {
	Describe("Validate", func() {
		It("returns nil when the constraint is valid", func() {
			con := constraint.Within(
				"ns",
				constraint.None,
				constraint.Not(
					constraint.And(
						constraint.Equal("a", "1"),
						constraint.NotEqual("b", "2"),
						constraint.Or(
							constraint.Empty("c"),
						),
					),
				),
			)

			err := con.Validate()

			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an error if a WITHIN constraint has an invalid namespace", func() {
			con := constraint.Within("_ns", constraint.Empty("a"))
			err := con.Validate()

			Expect(err).To(HaveOccurred())
		})

		It("returns an error if a WITHIN constraint has no terms", func() {
			con := constraint.Within("ns")
			err := con.Validate()

			Expect(err).To(HaveOccurred())
		})

		It("returns an error if a WITHIN constraint contains an invalid term", func() {
			con := constraint.Within("ns", constraint.And())
			err := con.Validate()

			Expect(err).To(HaveOccurred())
		})

		It("returns an error if a NOT constraint has an invalid term", func() {
			con := constraint.Not(constraint.And())
			err := con.Validate()

			Expect(err).To(HaveOccurred())
		})

		It("returns an error if an AND constraint has no terms", func() {
			con := constraint.And()
			err := con.Validate()

			Expect(err).To(HaveOccurred())
		})

		It("returns an error if an AND constraint has an invalid term", func() {
			con := constraint.And(constraint.And())
			err := con.Validate()

			Expect(err).To(HaveOccurred())
		})

		It("returns an error if an OR constraint has no terms", func() {
			con := constraint.Or()
			err := con.Validate()

			Expect(err).To(HaveOccurred())
		})

		It("returns an error if an OR constraint has an invalid term", func() {
			con := constraint.Or(constraint.And())
			err := con.Validate()

			Expect(err).To(HaveOccurred())
		})
	})

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
				"None",
				constraint.None,
				"{*}",
			),
			Entry(
				"None nested in compound constraint",
				constraint.And(
					constraint.None,
					constraint.None,
				),
				"{*, *}",
			),

			Entry(
				"Within",
				constraint.Within(
					"ns",
					constraint.Equal("a", "1"),
				),
				"ns::{a=1}",
			),
			Entry(
				"Within with multiple values",
				constraint.Within(
					"ns",
					constraint.Equal("a", "1"),
					constraint.Equal("b", "2"),
				),
				"ns::{a=1, b=2}",
			),

			Entry(
				"Equal",
				constraint.Equal("a", "1"),
				"{a=1}",
			),
			Entry(
				"NotEqual",
				constraint.NotEqual("a", "1"),
				"{a!=1}",
			),

			Entry(
				"Empty",
				constraint.Empty("a"),
				"{!a}",
			),
			Entry(
				"NotEmpty",
				constraint.NotEmpty("a"),
				"{a}",
			),

			Entry(
				"Not",
				constraint.Not(constraint.Equal("a", "1")),
				"{! a=1}",
			),
			Entry(
				"Not with compound expression",
				constraint.Not(
					constraint.And(
						constraint.Equal("a", "1"),
						constraint.Equal("b", "2"),
					),
				),
				"{! {a=1, b=2}}",
			),

			Entry(
				"And",
				constraint.And(
					constraint.Equal("a", "1"),
				),
				"{a=1}",
			),
			Entry(
				"And with multiple values",
				constraint.And(
					constraint.Equal("a", "1"),
					constraint.Equal("b", "2"),
				),
				"{a=1, b=2}",
			),

			Entry(
				"Or",
				constraint.Or(
					constraint.Equal("a", "1"),
				),
				"{a=1}",
			),
			Entry(
				"Or with multiple values",
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
