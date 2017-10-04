package attrmeta_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/constraint"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
)

var _ = Describe("Table", func() {
	Describe("MatchConstraint", func() {
		DescribeTable(
			"returns true when the table matches the constraint",
			func(t attrmeta.Table, ns string, con constraint.Constraint) {
				Expect(t.MatchConstraint(ns, con)).To(BeTrue())
			},

			Entry(
				"None",
				attrmeta.Table{},
				"ns",
				constraint.None,
			),

			Entry(
				"Within",
				attrmeta.Table{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.Within("ns", constraint.Equal("a", "1")),
			),

			Entry(
				"Equal",
				attrmeta.Table{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.Equal("a", "1"),
			),

			Entry(
				"NotEqual",
				attrmeta.Table{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.NotEqual("a", "2"),
			),

			Entry(
				"Not",
				attrmeta.Table{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.Not(constraint.Equal("a", "2")),
			),

			Entry(
				"And",
				attrmeta.Table{
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
				attrmeta.Table{
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
			func(t attrmeta.Table, ns string, con constraint.Constraint) {
				Expect(t.MatchConstraint(ns, con)).To(BeFalse())
			},

			Entry(
				"Within with failing constraint",
				attrmeta.Table{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.Within("ns", constraint.Equal("a", "2")),
			),
			Entry(
				"Within with different namespace",
				attrmeta.Table{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.Within("other", constraint.Equal("a", "1")),
			),

			Entry(
				"Equal",
				attrmeta.Table{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.Equal("a", "2"),
			),

			Entry(
				"NotEqual",
				attrmeta.Table{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.NotEqual("a", "1"),
			),

			Entry(
				"Not",
				attrmeta.Table{
					"ns": {"a": {Attr: rinq.Set("a", "1")}},
				},
				"ns",
				constraint.Not(constraint.Equal("a", "1")),
			),

			Entry(
				"And",
				attrmeta.Table{
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
				attrmeta.Table{
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

	Describe("CloneAndMerge", func() {
		var t attrmeta.Table

		BeforeEach(func() {
			t = attrmeta.Table{
				"ns1": {
					"a": {Attr: rinq.Set("a", "1")},
				},
				"ns2": {
					"b": {Attr: rinq.Set("b", "2")},
				},
			}
		})

		It("returns a different instance", func() {
			c := t.CloneAndMerge("ns2", attrmeta.Namespace{})

			c["ns3"] = attrmeta.Namespace{}
			Expect(t).NotTo(HaveKey("ns3"))
		})

		It("clones the contained namespaces", func() {
			c := t.CloneAndMerge("ns2", attrmeta.Namespace{})

			c["ns1"]["c"] = attrmeta.Attr{Attr: rinq.Set("c", "3")}
			Expect(t["ns1"]).NotTo(HaveKey("c"))
		})

		It("does not clone the merged namespace", func() {
			ns := attrmeta.Namespace{}

			c := t.CloneAndMerge("ns2", ns)

			c["ns2"]["c"] = attrmeta.Attr{Attr: rinq.Set("c", "3")}
			Expect(ns).To(HaveKey("c"))
		})

		It("replaces an existing namespace", func() {
			c := t.CloneAndMerge("ns2", attrmeta.Namespace{
				"c": {Attr: rinq.Set("c", "3")},
			})

			Expect(c).To(Equal(attrmeta.Table{
				"ns1": {
					"a": {Attr: rinq.Set("a", "1")},
				},
				"ns2": {
					"c": {Attr: rinq.Set("c", "3")},
				},
			}))
		})

		It("merges a new namespace", func() {
			c := t.CloneAndMerge("ns3", attrmeta.Namespace{
				"c": {Attr: rinq.Set("c", "3")},
			})

			Expect(c).To(Equal(attrmeta.Table{
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

	Describe("WriteTo", func() {
		buf := &bytes.Buffer{}

		BeforeEach(func() {
			buf.Reset()
		})

		Context("when the table is empty", func() {
			t := attrmeta.Table{}

			It("writes only braces", func() {
				t.WriteTo(buf)

				Expect(buf.String()).To(Equal("{}"))
			})
		})

		Context("when the table is not empty", func() {
			t := attrmeta.Table{
				"ns1": {
					"a": {Attr: rinq.Set("a", "1")},
				},
				"ns2": {
					"b": {Attr: rinq.Set("b", "2")},
				},
			}

			It("writes namespaces in any order", func() {
				var buf bytes.Buffer

				t.WriteTo(&buf)

				Expect(buf.String()).To(SatisfyAny(
					Equal("ns1::{a=1} ns2::{b=2}"),
					Equal("ns2::{b=2} ns1::{a=1}"),
				))
			})
		})

		It("excludes empty namespaces", func() {
			var buf bytes.Buffer

			t := attrmeta.Table{
				"ns1": {
					"a": {Attr: rinq.Set("a", "1")},
				},
				"ns2": {},
			}

			t.WriteTo(&buf)

			Expect(buf.String()).To(Equal("ns1::{a=1}"))
		})
	})

	Describe("String", func() {
		It("returns only braces when the table is empty", func() {
			Expect(attrmeta.Table{}.String()).To(Equal("{}"))
		})

		It("returns namespaces in any order", func() {
			t := attrmeta.Table{
				"ns1": {
					"a": {Attr: rinq.Set("a", "1")},
				},
				"ns2": {
					"b": {Attr: rinq.Set("b", "2")},
				},
			}

			Expect(t.String()).To(SatisfyAny(
				Equal("ns1::{a=1} ns2::{b=2}"),
				Equal("ns2::{b=2} ns1::{a=1}"),
			))
		})
	})
})
