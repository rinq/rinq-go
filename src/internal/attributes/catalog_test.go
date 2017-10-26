package attributes_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/constraint"
	. "github.com/rinq/rinq-go/src/internal/attributes"
)

var _ = Describe("Catalog", func() {
	Describe("WithNamespace", func() {
		var cat Catalog

		BeforeEach(func() {
			cat = Catalog{
				"ns1": {
					"a": {Attr: rinq.Set("a", "1")},
				},
				"ns2": {
					"b": {Attr: rinq.Set("b", "2")},
				},
			}
		})

		It("returns a different instance", func() {
			c := cat.WithNamespace("ns2", VTable{})

			c["ns3"] = VTable{}
			Expect(cat).NotTo(HaveKey("ns3"))
		})

		It("clones the contained namespaces", func() {
			c := cat.WithNamespace("ns2", VTable{})

			c["ns1"]["c"] = VAttr{Attr: rinq.Set("c", "3")}
			Expect(cat["ns1"]).NotTo(HaveKey("c"))
		})

		It("does not clone the merged namespace", func() {
			ns := VTable{}

			c := cat.WithNamespace("ns2", ns)

			c["ns2"]["c"] = VAttr{Attr: rinq.Set("c", "3")}
			Expect(ns).To(HaveKey("c"))
		})

		It("replaces an existing namespace", func() {
			c := cat.WithNamespace("ns2", VTable{
				"c": {Attr: rinq.Set("c", "3")},
			})

			Expect(c).To(Equal(Catalog{
				"ns1": {
					"a": {Attr: rinq.Set("a", "1")},
				},
				"ns2": {
					"c": {Attr: rinq.Set("c", "3")},
				},
			}))
		})

		It("merges a new namespace", func() {
			c := cat.WithNamespace("ns3", VTable{
				"c": {Attr: rinq.Set("c", "3")},
			})

			Expect(c).To(Equal(Catalog{
				"ns1": {
					"a": {Attr: rinq.Set("a", "1")},
				},
				"ns2": {
					"b": {Attr: rinq.Set("b", "2")},
				},
				"ns3": {
					"c": {Attr: rinq.Set("c", "3")},
				},
			}))
		})
	})

	Describe("MatchConstraint", func() {
		DescribeTable(
			"returns true when the catalog matches the constraint",
			func(cat Catalog, ns string, con constraint.Constraint) {
				Expect(cat.MatchConstraint(ns, con)).To(BeTrue())
			},

			Entry(
				"None",
				Catalog{},
				"ns",
				constraint.None,
			),

			Entry(
				"Within",
				Catalog{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.Within("ns", constraint.Equal("a", "1")),
			),

			Entry(
				"Equal",
				Catalog{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.Equal("a", "1"),
			),

			Entry(
				"NotEqual",
				Catalog{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.NotEqual("a", "2"),
			),

			Entry(
				"Not",
				Catalog{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.Not(constraint.Equal("a", "2")),
			),

			Entry(
				"And",
				Catalog{
					"ns": {
						"a": {Attr: rinq.Set("a", "1")},
						"b": {Attr: rinq.Set("b", "2")},
					},
				},
				"ns",
				constraint.And(
					constraint.Equal("a", "1"),
					constraint.Equal("b", "2"),
				),
			),

			Entry(
				"Or",
				Catalog{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.Or(
					constraint.Equal("a", "1"),
					constraint.Equal("a", "2"),
				),
			),
		)

		DescribeTable(
			"returns false when the table matches the constraint",
			func(cat Catalog, ns string, con constraint.Constraint) {
				Expect(cat.MatchConstraint(ns, con)).To(BeFalse())
			},

			Entry(
				"Within with failing constraint",
				Catalog{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.Within("ns", constraint.Equal("a", "2")),
			),
			Entry(
				"Within with different namespace",
				Catalog{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.Within("other", constraint.Equal("a", "1")),
			),

			Entry(
				"Equal",
				Catalog{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.Equal("a", "2"),
			),

			Entry(
				"NotEqual",
				Catalog{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.NotEqual("a", "1"),
			),

			Entry(
				"Not",
				Catalog{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.Not(constraint.Equal("a", "1")),
			),

			Entry(
				"And",
				Catalog{
					"ns": {
						"a": {Attr: rinq.Set("a", "1")},
						"b": {Attr: rinq.Set("b", "2")},
					},
				},
				"ns",
				constraint.And(
					constraint.Equal("a", "2"),
					constraint.Equal("b", "2"),
				),
			),

			Entry(
				"Or",
				Catalog{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.Or(
					constraint.Equal("a", "2"),
					constraint.Equal("a", "3"),
				),
			),
		)
	})

	Describe("IsEmpty", func() {
		It("returns true when the catalog is empty", func() {
			Expect(Catalog{}.IsEmpty()).To(BeTrue())
		})

		It("returns true when the catalog contains only empty namespaces", func() {
			Expect(Catalog{"ns": {}}.IsEmpty()).To(BeTrue())
		})

		It("returns false when the table is not empty", func() {
			cat := Catalog{
				"ns1": {
					"a": {Attr: rinq.Set("a", "1")},
				},
			}

			Expect(cat.IsEmpty()).To(BeFalse())
		})
	})

	Describe("String", func() {
		Context("when the table is empty", func() {
			cat := Catalog{}

			It("renders only braces", func() {
				Expect(cat.String()).To(Equal("{}"))
			})
		})

		Context("when the table is not empty", func() {
			It("writes namespaces in any order", func() {
				cat := Catalog{
					"ns1": {
						"a": {Attr: rinq.Set("a", "1")},
					},
					"ns2": {
						"b": {Attr: rinq.Set("b", "2")},
					},
				}

				Expect(cat.String()).To(SatisfyAny(
					Equal("ns1::{a=1} ns2::{b=2}"),
					Equal("ns2::{b=2} ns1::{a=1}"),
				))
			})
		})

		It("excludes empty namespaces", func() {
			cat := Catalog{
				"ns1": {
					"a": {Attr: rinq.Set("a", "1")},
				},
				"ns2": {},
			}

			Expect(cat.String()).To(Equal("ns1::{a=1}"))
		})
	})
})
