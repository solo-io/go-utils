package securityscanutils_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sort"

	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/securityscanutils"
)

const (
	glooRepoName = "gloo"
)

// In order to run these tests, you'll need to have trivy installed
// and on the PATH
var _ = Describe("Security Scan Suite", func() {
	var (
		outputDir string
	)

	BeforeEach(func() {
		checkTrivyInstall()
		var err error
		outputDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(outputDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Security Scanner", func() {
		It("works", func() {
			verConstraint, err := semver.NewConstraint("=v1.6.0 || =v1.7.0")
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("Output dir:", outputDir)
			secScanner := &SecurityScanner{
				Repos: []*SecurityScanRepo{{
					Repo:  glooRepoName,
					Owner: "solo-io",
					Opts: &SecurityScanOpts{
						OutputDir: outputDir,
						ImagesPerVersion: map[string][]string{
							"v1.6.0": {"gloo"},
							"v1.7.0": {"gloo", "discovery"},
						},
						VersionConstraint:      verConstraint,
						ImageRepo:              "quay.io/solo-io",
						UploadCodeScanToGithub: false,
					},
				}},
			}

			// Run security scan
			err = secScanner.GenerateSecurityScans(context.TODO())
			Expect(err).NotTo(HaveOccurred())

			ExpectDirToHaveFiles(outputDir, "gloo")
			// Have a directory for each repo we scanned
			markdownDir := path.Join(outputDir, "gloo", "markdown_results")
			// Have a directory for each version we scanned
			ExpectDirToHaveFiles(markdownDir, "1.6.0", "1.7.0")
			// Expect there to be a generated generated file for each image per version
			ExpectDirToHaveFiles(path.Join(markdownDir, "1.6.0"), "gloo_cve_report.docgen")
			ExpectDirToHaveFiles(path.Join(markdownDir, "1.7.0"), "discovery_cve_report.docgen", "gloo_cve_report.docgen")

			sarifDir := path.Join(outputDir, "gloo", "sarif_results")
			// Have a directory for each version we scanned
			ExpectDirToHaveFiles(sarifDir, "1.6.0", "1.7.0")
			// Expect there to be a generated sarif file for each image per version
			ExpectDirToHaveFiles(path.Join(sarifDir, "1.6.0"), "gloo_cve_report.sarif")
			ExpectDirToHaveFiles(path.Join(sarifDir, "1.7.0"), "discovery_cve_report.sarif", "gloo_cve_report.sarif")

		})

		It("errors if more than one constraint is matched", func() {
			verConstraint, err := semver.NewConstraint("=v1.7.0")
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("Output dir:", outputDir)
			secScanner := &SecurityScanner{
				Repos: []*SecurityScanRepo{{
					Repo:  glooRepoName,
					Owner: "solo-io",
					Opts: &SecurityScanOpts{
						OutputDir: outputDir,
						// Specify redundant constraints
						ImagesPerVersion: map[string][]string{
							"v1.7.0":   {"gloo", "discovery"},
							">=v1.7.0": {"gloo"},
						},
						VersionConstraint:      verConstraint,
						ImageRepo:              "quay.io/solo-io",
						UploadCodeScanToGithub: false,
					},
				}},
			}

			err = secScanner.GenerateSecurityScans(context.TODO())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("version 1.7.0 matched more than one constraint provided"))
		})

		It("errors if no constraint is matched", func() {
			verConstraint, err := semver.NewConstraint("=v1.7.0")
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("Output dir:", outputDir)
			secScanner := &SecurityScanner{
				Repos: []*SecurityScanRepo{{
					Repo:  glooRepoName,
					Owner: "solo-io",
					Opts: &SecurityScanOpts{
						OutputDir: outputDir,
						ImagesPerVersion: map[string][]string{
							"v1.6.0": {"gloo", "discovery"},
						},
						VersionConstraint:      verConstraint,
						ImageRepo:              "quay.io/solo-io",
						UploadCodeScanToGithub: false,
					},
				}},
			}

			err = secScanner.GenerateSecurityScans(context.TODO())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("version 1.7.0 matched no constraints and has no images to scan"))
		})
	})
})

// Trivy should be installed on PATH
func checkTrivyInstall() {
	path, err := exec.LookPath("trivy")
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, path).NotTo(BeEmpty())
}

// Accepts a list of file names and passes all tests only if the directory path passed in
// as dir includes all fileNames passed in.
func ExpectDirToHaveFiles(dir string, fileNames ...string) {
	dirResults, err := os.ReadDir(dir)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, dirResults).To(HaveLen(len(fileNames)))
	var dirFiles []string
	for _, res := range dirResults {
		dirFiles = append(dirFiles, res.Name())
	}
	sort.Strings(dirFiles)
	sort.Strings(fileNames)
	ExpectWithOffset(1, dirFiles).To(Equal(fileNames))
}
