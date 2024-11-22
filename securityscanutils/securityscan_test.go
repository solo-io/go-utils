package securityscanutils_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"

	"github.com/solo-io/go-utils/fileutils"

	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/securityscanutils"
)

const (
	repoName         = "gloo"
	gatewayOwnerName = "solo-io"
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
			verConstraint, err := semver.NewConstraint("=v1.14.0 || =v1.15.1")
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("Output dir:", outputDir)
			secScanner := &SecurityScanner{
				Repos: []*SecurityScanRepo{{
					Repo:  repoName,
					Owner: gatewayOwnerName,
					Opts: &SecurityScanOpts{
						OutputDir:           outputDir,
						OutputResultLocally: true,
						ImagesPerVersion: map[string][]string{
							"v1.14.0": {"gloo"},
							// Scan should continue in the case an image cannot be found
							"v1.15.1": {"thisimagecannotbefound", "gloo", "discovery"},
						},
						VersionConstraint: verConstraint,
						ImageRepo:         "quay.io/solo-io",
					},
				}},
			}

			// Run security scan
			err = secScanner.GenerateSecurityScans(context.TODO())
			Expect(err).NotTo(HaveOccurred())

			ExpectDirToHaveFiles(outputDir, "gloo")
			// Have a markdown file for each version we scanned
			glooDir := path.Join(outputDir, "gloo")
			ExpectDirToHaveFiles(glooDir, "issue_results", "markdown_results")
			githubIssueDir := path.Join(glooDir, "issue_results")
			ExpectDirToHaveFiles(githubIssueDir, "1.14.0.md", "1.15.1.md")
			// Have a directory for each repo we scanned
			markdownDir := path.Join(outputDir, "gloo", "markdown_results")
			// Have a directory for each version we scanned
			ExpectDirToHaveFiles(markdownDir, "1.14.0", "1.15.1")
			// Expect there to be a generated docgen file for each image per version
			ExpectDirToHaveFiles(path.Join(markdownDir, "1.14.0"), "gloo_cve_report.docgen")
			ExpectDirToHaveFiles(path.Join(markdownDir, "1.15.1"), "discovery_cve_report.docgen", "gloo_cve_report.docgen")
		})

		It("scans all images from all constraints matched", func() {
			verConstraint, err := semver.NewConstraint("=v1.15.0")
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("Output dir:", outputDir)
			secScanner := &SecurityScanner{
				Repos: []*SecurityScanRepo{{
					Repo:  repoName,
					Owner: gatewayOwnerName,
					Opts: &SecurityScanOpts{
						OutputDir: outputDir,
						// Specify redundant constraints
						ImagesPerVersion: map[string][]string{
							">v1.14.0":  {"gloo", "discovery"},
							">=v1.15.0": {"glooGreaterThan17"},
						},
						VersionConstraint: verConstraint,
						ImageRepo:         "quay.io/solo-io",
					},
				}},
			}

			imagesToScan, err := secScanner.Repos[0].GetImagesToScan(semver.MustParse("v1.15.7"))
			Expect(imagesToScan).To(ContainElements("gloo", "discovery", "glooGreaterThan17"))
		})

		It("errors if no constraint is matched", func() {
			verConstraint, err := semver.NewConstraint("=v1.15.0")
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("Output dir:", outputDir)
			secScanner := &SecurityScanner{
				Repos: []*SecurityScanRepo{{
					Repo:  repoName,
					Owner: gatewayOwnerName,
					Opts: &SecurityScanOpts{
						OutputDir: outputDir,
						ImagesPerVersion: map[string][]string{
							"v1.14.0": {"gloo", "discovery"},
						},
						VersionConstraint: verConstraint,
						ImageRepo:         "quay.io/solo-io",
					},
				}},
			}

			err = secScanner.GenerateSecurityScans(context.TODO())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("version 1.15.0 matched no constraints and has no images to scan"))
		})

		When("scan has unrecoverable error", func() {
			It("short-circuits", func() {
				verConstraint, err := semver.NewConstraint("=v1.13.0 || =v1.14.0")
				Expect(err).NotTo(HaveOccurred())
				fmt.Println("Output dir:", outputDir)
				secScanner := &SecurityScanner{
					Repos: []*SecurityScanRepo{{
						Repo:  repoName,
						Owner: gatewayOwnerName,
						Opts: &SecurityScanOpts{
							OutputDir:           outputDir,
							OutputResultLocally: true,
							ImagesPerVersion: map[string][]string{
								"v1.14.0": {"gloo; $(poorly formatted image name to force UnrecoverableError)"},
							},
							VersionConstraint: verConstraint,
							ImageRepo:         "quay.io/solo-io",
						},
					}},
				}

				// Run security scan
				err = secScanner.GenerateSecurityScans(context.TODO())
				Expect(err).To(MatchError(UnrecoverableErr))

				ExpectDirToHaveFiles(outputDir, "gloo")
				// No images scanned; no files should exist
				glooDir := path.Join(outputDir, "gloo")
				ExpectDirToHaveFiles(glooDir, "issue_results", "markdown_results")
				localIssueDir := path.Join(glooDir, "issue_results")
				ExpectDirToHaveFiles(localIssueDir)
				// Have a directory for each repo we scanned
				markdownDir := path.Join(outputDir, "gloo", "markdown_results")
				// Have a directory for each version we scanned
				ExpectDirToHaveFiles(markdownDir, "1.14.0")
				ExpectDirToHaveFiles(path.Join(markdownDir, "1.14.0"))
			})
		})

		When("scan has recoverable error", func() {
			It("contains error in generated file", func() {
				verConstraint, err := semver.NewConstraint("=v1.15.0")
				Expect(err).NotTo(HaveOccurred())
				fmt.Println("Output dir:", outputDir)
				secScanner := &SecurityScanner{
					Repos: []*SecurityScanRepo{{
						Repo:  repoName,
						Owner: gatewayOwnerName,
						Opts: &SecurityScanOpts{
							OutputDir:           outputDir,
							OutputResultLocally: true,
							ImagesPerVersion: map[string][]string{
								"v1.15.0": {"thisimagedoesnotexist"},
							},
							VersionConstraint: verConstraint,
							ImageRepo:         "quay.io/solo-io",
						},
					}},
				}

				// Run security scan
				err = secScanner.GenerateSecurityScans(context.TODO())
				Expect(err).NotTo(HaveOccurred())

				ExpectDirToHaveFiles(outputDir, "gloo")
				// No images scanned; no files should exist
				glooDir := path.Join(outputDir, "gloo")
				ExpectDirToHaveFiles(glooDir, "issue_results", "markdown_results")
				localIssueDir := path.Join(glooDir, "issue_results")
				ExpectDirToHaveFiles(localIssueDir, "1.15.0.md")
				contents, err := fileutils.ReadFileString(path.Join(localIssueDir, "1.15.0.md"))
				Expect(err).NotTo(HaveOccurred())
				Expect(contents).To(ContainSubstring(ImageNotFoundError.Error()))
				// Have a directory for each repo we scanned
				markdownDir := path.Join(outputDir, "gloo", "markdown_results")
				// Have a directory for each version we scanned
				ExpectDirToHaveFiles(markdownDir, "1.15.0")
				ExpectDirToHaveFiles(path.Join(markdownDir, "1.15.0"))
			})
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
