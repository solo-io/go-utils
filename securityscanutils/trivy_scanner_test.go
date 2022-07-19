package securityscanutils_test

import (
	"context"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/securityscanutils"
)

var _ = Describe("Trivy Scanner", func() {
	var (
		outputDir string
	)

	JustBeforeEach(func() {
	})

	JustAfterEach(func() {
		err := os.RemoveAll(outputDir)
		Expect(err).NotTo(HaveOccurred())
	})

	When("Vulnerabilities exist", func() {
		markdownTplFile, err := GetTemplateFile(MarkdownTrivyTemplate)
		Expect(err).NotTo(HaveOccurred())
		outputDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		t := NewTrivyScanner(nil)
		testImage := "quay.io/solo-io/gloo:1.11.1"
		a, b, c := t.ScanImage(context.TODO(), testImage, markdownTplFile, outputDir)
		a, b, c = a, b, c
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
