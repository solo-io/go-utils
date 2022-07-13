package securityscanutils_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Trivy Scanner", func() {

	var (
		outputDir string
	)

	BeforeEach(func() {
		var err error
		outputDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(outputDir)
		Expect(err).NotTo(HaveOccurred())
	})

	When("Vulnerabilities exist", func() {

	})

	When("Image is not found", func() {

	})

	When("Exec returns error", func() {

	})

	Context("Benchmark", func() {

		It("Should do repeated scans efficiently", func() {

		})
	})

})
