package rinq

import (
	"bytes"
	"log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("standardLogger", func() {
	Describe("Log", func() {
		It("logs to the internal logger", func() {
			var buffer bytes.Buffer
			parent := log.New(&buffer, "", 0)
			logger := &standardLogger{false, parent}

			logger.Log("pattern %s", "value")

			Expect(buffer.String()).To(Equal("pattern value\n"))
		})
	})

	Describe("IsDebug", func() {
		It("returns value of isDebug flag", func() {
			Expect(NewLogger(true).IsDebug()).To(BeTrue())
			Expect(NewLogger(false).IsDebug()).To(BeFalse())
		})
	})
})
