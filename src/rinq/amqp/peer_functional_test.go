// +build !without_amqp,!without_functests

package amqp_test

import (
	"context"
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq/internal/functest"
)

var _ = Describe("peer (functional)", func() {
	var ns string

	BeforeEach(func() {
		ns = functest.NewNamespace()
	})

	AfterEach(func() {
		functest.TearDownNamespaces()
	})

	Describe("ID", func() {
		It("returns a valid peer ID", func() {
			subject := functest.SharedPeer()

			err := subject.ID().Validate()
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Describe("Session", func() {
		It("returns a session that belongs to this peer", func() {
			subject := functest.SharedPeer()

			sess := subject.Session()
			defer sess.Destroy()

			Expect(sess.ID().Peer).To(Equal(subject.ID()))
		})

		It("returns a session with a non-zero seq component", func() {
			subject := functest.SharedPeer()

			sess := subject.Session()
			defer sess.Destroy()

			Expect(sess.ID().Seq).To(BeNumerically(">", 0))
		})

		It("returns a session even if the peer is stopped", func() {
			subject := functest.NewPeer()

			subject.Stop()
			<-subject.Done()

			sess := subject.Session()
			Expect(sess).ToNot(BeNil())

			sess.Destroy()
		})
	})

	Describe("Listen", func() {
		It("accepts command requests for the specified namespace", func() {
			subject := functest.SharedPeer()

			nonce := rand.Int63()
			err := subject.Listen(ns, functest.AlwaysReturn(nonce))
			Expect(err).Should(BeNil())

			sess := subject.Session()
			defer sess.Destroy()

			p, err := sess.Call(context.Background(), ns, "", nil)
			defer p.Close()

			Expect(err).ShouldNot(HaveOccurred())
			Expect(p.Value()).To(BeEquivalentTo(nonce))
		})

		It("does not accept command requests for other namespaces", func() {
			subject := functest.SharedPeer()

			err := subject.Listen(ns, functest.AlwaysPanic())
			Expect(err).Should(BeNil())

			sess := subject.Session()
			defer sess.Destroy()

			ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
			defer cancel()

			_, err = sess.Call(ctx, functest.NewNamespace(), "", nil)
			Expect(err).To(Equal(context.DeadlineExceeded))
		})

		It("changes the handler when invoked a second time", func() {
			subject := functest.SharedPeer()
			functest.Must(subject.Listen(ns, functest.AlwaysPanic()))

			nonce := rand.Int63()
			err := subject.Listen(ns, functest.AlwaysReturn(nonce))
			Expect(err).Should(BeNil())

			sess := subject.Session()
			defer sess.Destroy()

			p, err := sess.Call(context.Background(), ns, "", nil)
			defer p.Close()

			Expect(err).ShouldNot(HaveOccurred())
			Expect(p.Value()).To(BeEquivalentTo(nonce))
		})

		It("returns an error if the namespace is invalid", func() {
			subject := functest.SharedPeer()

			err := subject.Listen("_invalid", functest.AlwaysPanic())
			Expect(err).Should(HaveOccurred())
		})

		It("returns an error if the peer is stopped", func() {
			subject := functest.NewPeer()

			subject.Stop()
			<-subject.Done()

			err := subject.Listen(ns, functest.AlwaysPanic())
			Expect(err).Should(HaveOccurred())
		})
	})

	Describe("Unlisten", func() {
		It("stops accepting command requests", func() {
			subject := functest.SharedPeer()
			functest.Must(subject.Listen(ns, functest.AlwaysPanic()))

			err := subject.Unlisten(ns)
			Expect(err).Should(BeNil())

			sess := subject.Session()
			defer sess.Destroy()

			ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
			defer cancel()

			_, err = sess.Call(ctx, ns, "", nil)
			Expect(err).To(Equal(context.DeadlineExceeded))
		})

		It("can be invoked when not listening", func() {
			subject := functest.SharedPeer()

			err := subject.Unlisten("unused-namespace")
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("returns an error if the namespace is invalid", func() {
			subject := functest.SharedPeer()

			err := subject.Unlisten("_invalid")
			Expect(err).Should(HaveOccurred())
		})

		It("returns an error if the peer is stopped", func() {
			subject := functest.NewPeer()
			functest.Must(subject.Listen(ns, functest.AlwaysPanic()))

			subject.Stop()
			<-subject.Done()

			err := subject.Unlisten(ns)
			Expect(err).Should(HaveOccurred())
		})
	})

	Describe("Stop", func() {
		Context("when running normally", func() {
			It("cancels pending calls", func() {
				server := functest.SharedPeer()
				barrier := make(chan struct{})
				functest.Must(server.Listen(ns, functest.Barrier(barrier)))

				subject := functest.NewPeer()

				go func() {
					<-barrier
					subject.Stop()
					<-barrier
				}()

				sess := subject.Session()
				defer sess.Destroy()

				_, err := sess.Call(context.Background(), ns, "", nil)
				Expect(err).To(Equal(context.Canceled))
			})
		})

		Context("when stopping gracefully", func() {
			It("cancels pending calls", func() {
				server := functest.SharedPeer()
				barrier := make(chan struct{})
				functest.Must(server.Listen(ns, functest.Barrier(barrier)))

				subject := functest.NewPeer()

				go func() {
					<-barrier
					subject.GracefulStop()
					subject.Stop()
					<-barrier
				}()

				sess := subject.Session()
				defer sess.Destroy()

				_, err := sess.Call(context.Background(), ns, "", nil)
				Expect(err).To(Equal(context.Canceled))
			})
		})
	})

	Describe("GracefulStop", func() {
		It("waits for pending calls", func() {
			server := functest.SharedPeer()
			barrier := make(chan struct{})
			functest.Must(server.Listen(ns, functest.Barrier(barrier)))

			subject := functest.NewPeer()

			go func() {
				<-barrier
				subject.GracefulStop()
				<-barrier
			}()

			sess := subject.Session()
			defer sess.Destroy()

			_, err := sess.Call(context.Background(), ns, "", nil)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
