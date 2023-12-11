package securityscanutils_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"

	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/securityscanutils"
)

const (
	glooRepoName = "gloo"
)

var _ = Describe("Security Scan Suite", func() {

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
						OutputDir:           outputDir,
						OutputResultLocally: true,
						ImagesPerVersion: map[string][]string{
							"v1.6.0": {"gloo"},
							// Scan should continue in the case an image cannot be found
							"v1.7.0": {"thisimagecannotbefound", "gloo", "discovery"},
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
			// Have a markdown file for each version we scanned
			glooDir := path.Join(outputDir, "gloo")
			ExpectDirToHaveFiles(glooDir, "github_issue_results", "markdown_results")
			githubIssueDir := path.Join(glooDir, "github_issue_results")
			ExpectDirToHaveFiles(githubIssueDir, "1.6.0.md", "1.7.0.md")
			// Have a directory for each repo we scanned
			markdownDir := path.Join(outputDir, "gloo", "markdown_results")
			// Have a directory for each version we scanned
			ExpectDirToHaveFiles(markdownDir, "1.6.0", "1.7.0")
			// Expect there to be a generated docgen file for each image per version
			ExpectDirToHaveFiles(path.Join(markdownDir, "1.6.0"), "gloo_cve_report.docgen")
			ExpectDirToHaveFiles(path.Join(markdownDir, "1.7.0"), "discovery_cve_report.docgen", "gloo_cve_report.docgen")
		})

		It("scans all images from all constraints matched", func() {
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
							">v1.6.0":  {"gloo", "discovery"},
							">=v1.7.0": {"glooGreaterThan17"},
						},
						VersionConstraint:      verConstraint,
						ImageRepo:              "quay.io/solo-io",
						UploadCodeScanToGithub: false,
					},
				}},
			}

			imagesToScan, err := secScanner.Repos[0].GetImagesToScan(semver.MustParse("v1.7.7"))
			Expect(imagesToScan).To(ContainElements("gloo", "discovery", "glooGreaterThan17"))
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

// Accepts a list of file names and passes all tests only if the directory path passed in
// as dir includes all fileNames passed in.
func ExpectDirToHaveFiles(dir string, fileNames ...string) {
	dirResults, err := ioutil.ReadDir(dir)
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
