package securityscanutils_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/osutils/executils"
	. "github.com/solo-io/go-utils/securityscanutils"
)

var _ = Describe("Trivy Scanner", func() {
	var (
		t                         *TrivyScanner
		inputImage                string
		inputMarkdownTemplateFile string
		outputDir, outputFile     string
		err                       error
	)

	JustBeforeEach(func() {
		t = NewTrivyScanner(executils.CombinedOutputWithStatus)
		inputMarkdownTemplateFile, err = GetTemplateFile(MarkdownTrivyTemplate)
		Expect(err).NotTo(HaveOccurred())
		outputDir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		outputFile = filepath.Join(outputDir, "test_report.docgen")
	})

	JustAfterEach(func() {
		err := os.RemoveAll(outputDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("Finds vulnerabilities", func() {
		inputImage = "quay.io/solo-io/gloo:1.11.1"
		completed, vulnFound, err := t.ScanImage(context.TODO(), inputImage, inputMarkdownTemplateFile, outputFile)

		Expect(err).NotTo(HaveOccurred())
		Expect(completed).To(Equal(true))
		Expect(vulnFound).To(Equal(true))
	})

	It("Cannot find Image", func() {
		inputImage = "quay.io/solo-io/gloo:1.11.13245"
		completed, vulnFound, err := t.ScanImage(context.TODO(), inputImage, inputMarkdownTemplateFile, outputFile)

		Expect(err).NotTo(HaveOccurred())
		Expect(completed).To(Equal(false))
		Expect(vulnFound).To(Equal(false))

	})

	It("Returns error from Exec via Timeout", func() {
		inputImage = "quay.io/solo-io/gloo:1.11.1"
		completed, vulnFound, err := t.ScanImage(context.TODO(), "", "", "")

		//Error occurs when all trivy scan arguments are empty
		Expect(err).To(HaveOccurred())
		Expect(completed).To(Equal(false))
		Expect(vulnFound).To(Equal(false))

	})

	FContext("Benchmark", func() {
		It("Should do repeated scans efficiently", func() {
			inputImage = "quay.io/solo-io/gloo:1.11.1"
			attemptStart := time.Now()
			for i := 0; i < 10; i++ {
				_, _, err := t.ScanImage(context.TODO(), inputImage, inputMarkdownTemplateFile, outputFile)
				Expect(err).NotTo(HaveOccurred())
			}
			attemptEnd := time.Since(attemptStart)
			Expect(attemptEnd).To(BeNumerically("<", 20*time.Second))
		})
	})
})
