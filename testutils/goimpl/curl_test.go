package goimpl

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Curl", func() {

	It("should query a url", func() {
		soloResp, err := Curl("https://www.solo.io")
		Expect(err).NotTo(HaveOccurred())
		Expect(len(soloResp)).To(BeNumerically(">", 0))
	})

	It("should error on invalid url", func() {
		_, err := Curl("invalid://www.solo.io")
		Expect(err).To(HaveOccurred())
	})

})
