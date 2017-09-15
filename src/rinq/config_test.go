package rinq_test

import (
	"os"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rinq/rinq-go/src/rinq"
)

var _ = Describe("DefaultConfig", func() {
	It("uses the correct defaults", func() {
		Expect(rinq.DefaultConfig).To(Equal(rinq.Config{
			DefaultTimeout: 5 * time.Second,
			CommandWorkers: runtime.GOMAXPROCS(0),
			SessionWorkers: runtime.GOMAXPROCS(0) * 10,
			Logger:         rinq.NewLogger(false),
			PruneInterval:  3 * time.Minute,
			Product:        "",
		}))
	})
})

var _ = Describe("ConfigFromEnv", func() {
	AfterEach(func() {
		os.Setenv("RINQ_DEFAULT_TIMEOUT", "")
		os.Setenv("RINQ_LOG_DEBUG", "")
		os.Setenv("RINQ_COMMAND_WORKERS", "")
		os.Setenv("RINQ_SESSION_WORKERS", "")
		os.Setenv("RINQ_PRUNE_INTERVAL", "")
		os.Setenv("RINQ_PRODUCT", "")
	})

	Context("RINQ_DEFAULT_TIMEOUT", func() {
		It("populates the DefaultTimeout field in milliseconds", func() {
			os.Setenv("RINQ_DEFAULT_TIMEOUT", "500")
			cfg, err := rinq.NewConfigFromEnv()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.DefaultTimeout).To(Equal(500 * time.Millisecond))
		})

		It("uses the default value when undefined", func() {
			cfg, err := rinq.NewConfigFromEnv()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.DefaultTimeout).To(Equal(rinq.DefaultConfig.DefaultTimeout))
		})

		It("returns an error if the value is not a positive integer", func() {
			os.Setenv("RINQ_DEFAULT_TIMEOUT", "-500")
			_, err := rinq.NewConfigFromEnv()

			Expect(err).To(HaveOccurred())
		})
	})

	Context("RINQ_LOG_DEBUG", func() {
		It("uses a debug logger when set to true", func() {
			os.Setenv("RINQ_LOG_DEBUG", "true")
			cfg, err := rinq.NewConfigFromEnv()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Logger).To(Equal(rinq.NewLogger(true)))
		})

		It("uses a non-debug logger when set to false", func() {
			os.Setenv("RINQ_LOG_DEBUG", "false")
			cfg, err := rinq.NewConfigFromEnv()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Logger).To(Equal(rinq.NewLogger(false)))
		})

		It("uses the default when undefined", func() {
			cfg, err := rinq.NewConfigFromEnv()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Logger).To(Equal(rinq.DefaultConfig.Logger))
		})

		It("returns an error if the value is not a boolean", func() {
			os.Setenv("RINQ_LOG_DEBUG", "invalid")
			_, err := rinq.NewConfigFromEnv()

			Expect(err).To(HaveOccurred())
		})
	})

	Context("RINQ_COMMAND_WORKERS", func() {
		It("populates the CommandWorkers field", func() {
			os.Setenv("RINQ_COMMAND_WORKERS", "15")
			cfg, err := rinq.NewConfigFromEnv()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.CommandWorkers).To(Equal(15))
		})

		It("uses the default value when undefined", func() {
			cfg, err := rinq.NewConfigFromEnv()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.CommandWorkers).To(Equal(rinq.DefaultConfig.CommandWorkers))
		})

		It("returns an error if the value is not a positive integer", func() {
			os.Setenv("RINQ_COMMAND_WORKERS", "-500")
			_, err := rinq.NewConfigFromEnv()

			Expect(err).To(HaveOccurred())
		})
	})

	Context("RINQ_SESSION_WORKERS", func() {
		It("populates the SessionWorkers field", func() {
			os.Setenv("RINQ_SESSION_WORKERS", "25")
			cfg, err := rinq.NewConfigFromEnv()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.SessionWorkers).To(Equal(25))
		})

		It("uses the default value when undefined", func() {
			cfg, err := rinq.NewConfigFromEnv()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.SessionWorkers).To(Equal(rinq.DefaultConfig.SessionWorkers))
		})

		It("returns an error if the value is not a positive integer", func() {
			os.Setenv("RINQ_SESSION_WORKERS", "-500")
			_, err := rinq.NewConfigFromEnv()

			Expect(err).To(HaveOccurred())
		})
	})

	Context("RINQ_PRUNE_INTERVAL", func() {
		It("populates the PruneInterval field in milliseconds", func() {
			os.Setenv("RINQ_PRUNE_INTERVAL", "1500")
			cfg, err := rinq.NewConfigFromEnv()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.PruneInterval).To(Equal(1500 * time.Millisecond))
		})

		It("uses the default value when undefined", func() {
			cfg, err := rinq.NewConfigFromEnv()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.PruneInterval).To(Equal(rinq.DefaultConfig.PruneInterval))
		})

		It("returns an error if the value is not a positive integer", func() {
			os.Setenv("RINQ_PRUNE_INTERVAL", "-500")
			_, err := rinq.NewConfigFromEnv()

			Expect(err).To(HaveOccurred())
		})
	})

	Context("RINQ_PRODUCT", func() {
		It("populates the Product field", func() {
			os.Setenv("RINQ_PRODUCT", "my-app")
			cfg, err := rinq.NewConfigFromEnv()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Product).To(Equal("my-app"))
		})

		It("uses the default value when undefined", func() {
			cfg, err := rinq.NewConfigFromEnv()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Product).To(Equal(rinq.DefaultConfig.Product))
		})
	})
})
