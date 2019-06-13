package debugutils

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("logs", func() {
	FContext("request builder", func() {
		var (
			requestBuilder *LogRequestBuilder
		)
		BeforeEach(func() {
			var err error
			requestBuilder, err = NewLogRequestBuilder()
			Expect(err).NotTo(HaveOccurred())
		})
		It("can properly build the requests from the gloo manifest", func() {
			requests, err := requestBuilder.LogsFromManifest(manifests)
			Expect(err).NotTo(HaveOccurred())
			Expect(requests).To(HaveLen(4))
		})
	})
})