package rinq_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("Constraint", func() {
	Describe("And", func() {
		It("returns an AND constraint", func() {
			e := rinq.Equal("a", "1").And(
				rinq.Equal("b", "2"),
			)

			Expect(e.String()).To(Equal(`{a=1, b=2}`))
		})
	})

	Describe("Or", func() {
		It("returns an OR constraint", func() {
			e := rinq.Equal("a", "1").Or(
				rinq.Equal("b", "2"),
			)

			Expect(e.String()).To(Equal(`{a=1|b=2}`))
		})
	})
})

var _ = Describe("Within", func() {
	It("returns the expected constraint", func() {
		e := rinq.Within(
			"ns",
			rinq.Equal("a", "1"),
		)

		Expect(e.String()).To(Equal(`ns::{a=1}`))
	})

	It("returns the expected constraint when multiple constraints are passed", func() {
		e := rinq.Within(
			"ns",
			rinq.Equal("a", "1"),
			rinq.Equal("b", "2"),
		)

		Expect(e.String()).To(Equal(`ns::{a=1, b=2}`))
	})
})

var _ = Describe("Equal", func() {
	It("returns the expected constraint", func() {
		e := rinq.Equal("a", "1")

		Expect(e.String()).To(Equal(`a=1`))
	})

	It("returns the expected constraint when multiple values are passed", func() {
		e := rinq.Equal("a", "1", "2")

		Expect(e.String()).To(Equal(`a IN (1, 2)`))
	})
})

var _ = Describe("NotEqual", func() {
	It("returns the expected constraint", func() {
		e := rinq.NotEqual("a", "1")

		Expect(e.String()).To(Equal(`a!=1`))
	})

	It("returns the expected constraint when multiple values are passed", func() {
		e := rinq.NotEqual("a", "1", "2")

		Expect(e.String()).To(Equal(`a NOT IN (1, 2)`))
	})
})

var _ = Describe("Empty", func() {
	It("returns the expected constraint", func() {
		e := rinq.Empty("a")

		Expect(e.String()).To(Equal(`!a`))
	})
})

var _ = Describe("NotEmpty", func() {
	It("returns the expected constraint", func() {
		e := rinq.NotEmpty("a")

		Expect(e.String()).To(Equal(`a`))
	})
})

var _ = Describe("Not", func() {
	It("returns the expected constraint", func() {
		e := rinq.Not(
			rinq.Equal("a", "1"),
		)

		Expect(e.String()).To(Equal(`NOT a=1`))
	})
})

var _ = Describe("And", func() {
	It("returns the expected constraint", func() {
		e := rinq.And(
			rinq.Equal("a", "1"),
			rinq.Equal("b", "2"),
		)

		Expect(e.String()).To(Equal(`{a=1, b=2}`))
	})

	It("returns the string representation of a single child", func() {
		e := rinq.And(
			rinq.Equal("a", "1"),
		)

		Expect(e.String()).To(Equal(`a=1`))
	})
})

var _ = Describe("Or", func() {
	It("returns the expected constraint", func() {
		e := rinq.Or(
			rinq.Equal("a", "1"),
			rinq.Equal("b", "2"),
		)

		Expect(e.String()).To(Equal(`{a=1|b=2}`))
	})

	It("returns the string representation of a single child", func() {
		e := rinq.Or(
			rinq.Equal("a", "1"),
		)

		Expect(e.String()).To(Equal(`a=1`))
	})
})
