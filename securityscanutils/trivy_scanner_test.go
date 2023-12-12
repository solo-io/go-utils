package securityscanutils_test

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rotisserie/eris"

	. "github.com/onsi/ginkgo/v2"
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
		inputImage = "quay.io/solo-io/gloo:1.11.1"
	})

	JustAfterEach(func() {
		err := os.RemoveAll(outputDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("Finds vulnerabilities", func() {
		t = NewTrivyScanner(func(cmd *exec.Cmd) ([]byte, int, error) {
			return nil, VulnerabilityFoundStatusCode, nil
		})
		completed, vulnFound, err := t.ScanImage(context.TODO(), inputImage, inputMarkdownTemplateFile, outputFile)

		Expect(err).NotTo(HaveOccurred())
		Expect(completed).To(Equal(true))
		Expect(vulnFound).To(Equal(true))
	})

	It("Cannot find Image", func() {
		t = NewTrivyScanner(func(cmd *exec.Cmd) ([]byte, int, error) {
			return []byte("Error containing the string 'No such image: '"), VulnerabilityFoundStatusCode + 1, eris.Errorf("Unread error")
		})
		completed, vulnFound, err := t.ScanImage(context.TODO(), inputImage, inputMarkdownTemplateFile, outputFile)

		Expect(err).To(MatchError(imageNotFoundError))
		Expect(completed).To(Equal(false))
		Expect(vulnFound).To(Equal(false))

	})

	It("Returns error from Exec via Timeout", func() {
		t = NewTrivyScanner(func(cmd *exec.Cmd) ([]byte, int, error) {
			return nil, VulnerabilityFoundStatusCode + 1, eris.Errorf("Unread error")
		})
		completed, vulnFound, err := t.ScanImage(context.TODO(), inputImage, inputMarkdownTemplateFile, outputFile)

		//Error occurs when all trivy scan arguments are empty
		Expect(err).To(HaveOccurred())
		Expect(completed).To(Equal(false))
		Expect(vulnFound).To(Equal(false))

	})

	It("Returns the correct status code with mock executor returning no errors", func() {
		MockCmdExecutor := func(cmd *exec.Cmd) ([]byte, int, error) {
			return nil, VulnerabilityFoundStatusCode + 1, nil
		}
		tMock := NewTrivyScanner(MockCmdExecutor)
		completed, vulnFound, err := tMock.ScanImage(context.TODO(), inputImage, inputMarkdownTemplateFile, outputFile)

		Expect(err).To(BeNil())
		Expect(completed).To(Equal(true))
		Expect(vulnFound).To(Equal(false))
	})

	It("Times out while backing off and retrying when mock executor returns error", func() {
		MockCmdExecutor := func(cmd *exec.Cmd) ([]byte, int, error) {
			return nil, VulnerabilityFoundStatusCode + 1, eris.Errorf("This is a fake error")
		}
		tMock := NewTrivyScanner(MockCmdExecutor)
		completed, vulnFound, err := tMock.ScanImage(context.TODO(), inputImage, inputMarkdownTemplateFile, outputFile)

		Expect(err).To(HaveOccurred())
		Expect(completed).To(Equal(false))
		Expect(vulnFound).To(Equal(false))
	})

	Context("Trivy Integration Tests", func() {
		It("Should do repeated scans without flaking", func() {
			inputImage = "quay.io/solo-io/gloo:1.11.1"
			samples := 8
			for i := 0; i < samples; i++ {
				_, _, err := t.ScanImage(context.TODO(), inputImage, inputMarkdownTemplateFile, outputFile)
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})
})
