package rinq

import (
	"bytes"
	"log"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = Describe("standardLogger", func() {
	Describe("Log", func() {
		It("logs to the internal logger", func() {
			var buffer bytes.Buffer
			parent := log.New(&buffer, "", 0)
			logger := &standardLogger{false, parent}

			logger.Log("pattern %s", "value")

			gomega.Expect(buffer.String()).To(gomega.Equal("pattern value\n"))
		})
	})

	Describe("IsDebug", func() {
		It("returns value of isDebug flag", func() {
			gomega.Expect(NewLogger(true).IsDebug()).To(gomega.BeTrue())
			gomega.Expect(NewLogger(false).IsDebug()).To(gomega.BeFalse())
		})
	})
})
