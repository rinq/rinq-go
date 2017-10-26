// +build !without_amqp,!without_functests

package remotesession_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/internal/attributes"
	"github.com/rinq/rinq-go/src/rinq/internal/functest"
)

var _ = Describe("revision (functional)", func() {
	var (
		ctx            context.Context
		ns             string
		client, server rinq.Peer
		session        rinq.Session
		local, remote  rinq.Revision
	)

	BeforeEach(func() {
		ctx = context.Background()
		ns = functest.NewNamespace()
		client = functest.NewPeer()
		session = client.Session()
		server = functest.NewPeer()

		functest.Must(server.Listen(ns, func(ctx context.Context, req rinq.Request, res rinq.Response) {
			remote = req.Source
			res.Close()
		}))

		local, _ = session.CurrentRevision()
		functest.Must(session.Call(ctx, ns, "", nil))
	})

	AfterEach(func() {
		functest.TearDownNamespaces()

		client.Stop()
		server.Stop()

		<-client.Done()
		<-server.Done()
	})

	Describe("Ref", func() {
		It("returns the same ref as the local revision", func() {
			Expect(remote.Ref()).To(Equal(local.Ref()))
		})
	})

	Describe("Refresh", func() {
		It("returns a revision with the same ref as the lastest local revision", func() {
			var err error
			local, err = local.Update(ctx, ns, rinq.Set("a", "1"))
			Expect(err).NotTo(HaveOccurred())

			remote, err = remote.Refresh(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(remote.Ref()).To(Equal(local.Ref()))
		})

		It("returns a not found error if the session has been destroyed", func() {
			session.Destroy()

			_, err := remote.Refresh(ctx)
			Expect(err).To(HaveOccurred())
			Expect(rinq.IsNotFound(err)).To(BeTrue())
		})
	})

	Describe("Get", func() {
		It("returns an empty attribute at revision zero", func() {
			attr, err := remote.Get(ctx, ns, "a")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr).To(Equal(rinq.Set("a", "")))
		})

		It("returns an empty attribute at revision zero after the session is destroyed", func() {
			session.Destroy()

			attr, err := remote.Get(ctx, ns, "a")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr).To(Equal(rinq.Set("a", "")))
		})

		It("returns an empty attribute when none exists", func() {
			var err error
			local, err = local.Update(ctx, ns, rinq.Set("a", "1"))
			Expect(err).NotTo(HaveOccurred())

			remote, err = remote.Refresh(ctx)
			Expect(err).NotTo(HaveOccurred())

			attr, err := remote.Get(ctx, ns, "b")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr).To(Equal(rinq.Set("b", "")))
		})

		It("returns an attribute created on the owning peer", func() {
			var err error
			local, err = local.Update(ctx, ns, rinq.Set("a", "1"))
			Expect(err).NotTo(HaveOccurred())

			remote, err = remote.Refresh(ctx)
			Expect(err).NotTo(HaveOccurred())

			attr, err := remote.Get(ctx, ns, "a")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr).To(Equal(rinq.Set("a", "1")))
		})

		It("returns an attribute updated on the remote peer from the cache", func() {
			// setup a handler that updates an attribute remotely
			functest.Must(server.Listen(ns, func(ctx context.Context, req rinq.Request, res rinq.Response) {
				var err error
				remote, err = req.Source.Update(ctx, ns, rinq.Set("a", "1"))
				Expect(err).NotTo(HaveOccurred())
				res.Close()
			}))

			// invoke the handler
			_, err := session.Call(ctx, ns, "", nil)
			Expect(err).NotTo(HaveOccurred())

			local, err = session.CurrentRevision()
			Expect(err).NotTo(HaveOccurred())

			// update the attribute locally such that the server does not know
			// about the change
			local, err = local.Update(ctx, ns, rinq.Set("a", "2"))
			Expect(err).NotTo(HaveOccurred())

			// the remote revision from the command handler should still be able
			// too pull the first value from its cache
			attr, err := remote.Get(ctx, ns, "a")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.Value).To(Equal("1"))
		})

		It("returns a stale fetch error if the attribute has been updated in a later revision", func() {
			var err error
			local, err = local.Update(ctx, ns, rinq.Set("a", "1"))
			Expect(err).NotTo(HaveOccurred())

			remote, err = remote.Refresh(ctx)
			Expect(err).NotTo(HaveOccurred())

			local, err = local.Update(ctx, ns, rinq.Set("a", "2"))
			Expect(err).NotTo(HaveOccurred())

			_, err = remote.Get(ctx, ns, "a")
			Expect(err).To(HaveOccurred())
			Expect(rinq.ShouldRetry(err)).To(BeTrue())
		})

		It("returns a not found error if the session has been destroyed", func() {
			// bump the version otherwise Get knows to return an empty attribute
			// for revision zero.
			var err error
			local, err = local.Update(ctx, ns, rinq.Set("a", "1"))
			Expect(err).NotTo(HaveOccurred())

			remote, err = remote.Refresh(ctx)
			Expect(err).NotTo(HaveOccurred())

			session.Destroy()

			_, err = remote.Get(ctx, ns, "a")
			Expect(err).To(HaveOccurred())
			Expect(rinq.IsNotFound(err)).To(BeTrue())
		})
	})

	Describe("GetMany", func() {
		It("returns empty attributes at revision zero", func() {
			attrs, err := remote.GetMany(ctx, ns, "a", "b")
			Expect(err).NotTo(HaveOccurred())
			Expect(attributes.ToMap(attrs)).To(Equal(
				map[string]rinq.Attr{
					"a": {Key: "a"},
					"b": {Key: "b"},
				},
			))
		})

		It("returns empty attributes at revision zero after the session is destroyed", func() {
			session.Destroy()

			attrs, err := remote.GetMany(ctx, ns, "a", "b")
			Expect(err).NotTo(HaveOccurred())
			Expect(attributes.ToMap(attrs)).To(Equal(
				map[string]rinq.Attr{
					"a": {Key: "a"},
					"b": {Key: "b"},
				},
			))
		})

		It("returns empty attributes when none exist", func() {
			var err error
			local, err = local.Update(ctx, ns, rinq.Set("a", "1"))
			Expect(err).NotTo(HaveOccurred())

			remote, err = remote.Refresh(ctx)
			Expect(err).NotTo(HaveOccurred())

			attrs, err := remote.GetMany(ctx, ns, "b", "c")
			Expect(err).NotTo(HaveOccurred())
			Expect(attributes.ToMap(attrs)).To(Equal(
				map[string]rinq.Attr{
					"b": {Key: "b"},
					"c": {Key: "c"},
				},
			))
		})

		It("returns an empty attribute table when called without keys", func() {
			session.Destroy() // even if the session has been destroyed

			attrs, err := remote.GetMany(ctx, ns)
			Expect(err).NotTo(HaveOccurred())
			Expect(attrs.IsEmpty()).To(BeTrue())
		})

		It("returns attributes created on the owning peer", func() {
			var err error
			local, err = local.Update(
				ctx,
				ns,
				rinq.Set("a", "1"),
				rinq.Set("b", "2"),
			)
			Expect(err).NotTo(HaveOccurred())

			remote, err = remote.Refresh(ctx)
			Expect(err).NotTo(HaveOccurred())

			attrs, err := remote.GetMany(ctx, ns, "a", "b")
			Expect(err).NotTo(HaveOccurred())
			Expect(attributes.ToMap(attrs)).To(Equal(
				map[string]rinq.Attr{
					"a": rinq.Set("a", "1"),
					"b": rinq.Set("b", "2"),
				},
			))
		})

		It("returns a stale fetch error if the attribute has been updated in a later revision", func() {
			var err error
			local, err = local.Update(ctx, ns, rinq.Set("a", "1"))
			Expect(err).NotTo(HaveOccurred())

			remote, err = remote.Refresh(ctx)
			Expect(err).NotTo(HaveOccurred())

			local, err = local.Update(ctx, ns, rinq.Set("a", "2"))
			Expect(err).NotTo(HaveOccurred())

			_, err = remote.GetMany(ctx, ns, "a")
			Expect(err).To(HaveOccurred())
			Expect(rinq.ShouldRetry(err)).To(BeTrue())
		})

		It("returns a not found error if the session has been destroyed", func() {
			// bump the version otherwise Get knows to return an empty attribute
			// for revision zero.
			var err error
			local, err = local.Update(ctx, ns, rinq.Set("a", "1"))
			Expect(err).NotTo(HaveOccurred())

			remote, err = remote.Refresh(ctx)
			Expect(err).NotTo(HaveOccurred())

			session.Destroy()

			_, err = remote.GetMany(ctx, ns, "a")
			Expect(err).To(HaveOccurred())
			Expect(rinq.IsNotFound(err)).To(BeTrue())
		})
	})

	Describe("Update", func() {
		It("returns a stale update error if session is at a later revision", func() {
			var err error
			local, err = local.Update(ctx, ns, rinq.Set("a", "1"))
			Expect(err).NotTo(HaveOccurred())

			_, err = remote.Update(ctx, ns, rinq.Set("a", "2"))
			Expect(err).To(HaveOccurred())
			Expect(rinq.ShouldRetry(err)).To(BeTrue())
		})

		It("returns a not found error if the session has been destroyed", func() {
			session.Destroy()

			var err error
			remote, err = remote.Update(ctx, ns, rinq.Set("a", "1"))
			Expect(err).To(HaveOccurred())
			Expect(rinq.IsNotFound(err)).To(BeTrue())
		})
	})

	Describe("Clear", func() {
		It("clears the attributes", func() {
			var err error
			local, err = local.Update(
				ctx,
				ns,
				rinq.Set("a", "1"),
				rinq.Set("b", "2"),
			)
			Expect(err).NotTo(HaveOccurred())

			remote, err = remote.Refresh(ctx)
			Expect(err).NotTo(HaveOccurred())

			_, err = remote.Clear(ctx, ns)
			Expect(err).NotTo(HaveOccurred())

			local, err = local.Refresh(ctx)
			Expect(err).NotTo(HaveOccurred())

			attrs, err := local.GetMany(ctx, ns, "a", "b")
			Expect(err).NotTo(HaveOccurred())

			a, _ := attrs.Get("a")
			b, _ := attrs.Get("b")
			Expect(a.Value).To(BeEmpty())
			Expect(b.Value).To(BeEmpty())
		})

		It("returns an error if any frozen attribute exists", func() {
			var err error
			local, err = local.Update(ctx, ns, rinq.Freeze("a", "1"))
			Expect(err).NotTo(HaveOccurred())

			remote, err = remote.Refresh(ctx)
			Expect(err).NotTo(HaveOccurred())

			_, err = remote.Clear(ctx, ns)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(rinq.FrozenAttributesError{Ref: remote.Ref()}))
		})

		It("returns a stale update error if session is at a later revision", func() {
			var err error
			local, err = local.Update(ctx, ns, rinq.Set("a", "1"))
			Expect(err).NotTo(HaveOccurred())

			_, err = remote.Clear(ctx, ns)
			Expect(err).To(HaveOccurred())
			Expect(rinq.ShouldRetry(err)).To(BeTrue())
		})

		It("returns a not found error if the session has been destroyed", func() {
			session.Destroy()

			var err error
			remote, err = remote.Clear(ctx, ns)
			Expect(err).To(HaveOccurred())
			Expect(rinq.IsNotFound(err)).To(BeTrue())
		})
	})

	Describe("Destroy", func() {
		It("returns a stale update error if session is at a later revision", func() {
			var err error
			local, err = local.Update(ctx, ns, rinq.Set("a", "1"))
			Expect(err).NotTo(HaveOccurred())

			err = remote.Destroy(ctx)
			Expect(err).To(HaveOccurred())
			Expect(rinq.ShouldRetry(err)).To(BeTrue())
		})

		It("returns a not found error if the session has been destroyed", func() {
			session.Destroy()

			err := remote.Destroy(ctx)
			Expect(err).To(HaveOccurred())
			Expect(rinq.IsNotFound(err)).To(BeTrue())
		})
	})
})
