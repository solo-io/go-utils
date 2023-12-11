package securityscanutils

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/solo-io/go-utils/osutils/executils"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/Masterminds/semver/v3"
	"github.com/solo-io/go-utils/stringutils"

	"github.com/google/go-github/v32/github"
	"github.com/imroc/req"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/githubutils"
)

type SecurityScanner struct {
	Repos        []*SecurityScanRepo
	githubClient *github.Client
}

type SecurityScanRepo struct {
	Repo  string
	Owner string
	Opts  *SecurityScanOpts

	// A set of private properties that are not constructed by the user
	releasesToScan []*github.RepositoryRelease

	// The RepositoryReleasePredicate used to determine if a particular release
	// should be run through our scanner
	scanReleasePredicate githubutils.RepositoryReleasePredicate

	trivyScanner *TrivyScanner

	// The writer responsible for generating Github Issues for certain releases
	githubIssueWriter *GithubIssueWriter
}

type SecurityScanOpts struct {
	// The following directory structure will be created in your output dir.
	/*
	   OUTPUT_DIR/
	   ├─ markdown_results/
	   │  ├─ repo1/
	   │  │  ├─ 1.4.12/
	   │  │  ├─ 1.5.0/
	   │  ├─ repo2/
	   │  │  ├─ 1.4.13/
	   │  │  ├─ 1.5.1/
	   ├─ sarif_results/
	   │  ├─ repo1/
	   │  │  ├─ 1.4.12/
	   │  │  ├─ 1.5.0/
	   │  ├─ repo2/
	   │  │  ├─ 1.4.13/
	   │  │  ├─ 1.5.1/
	*/
	OutputDir string
	// A mapping of version constraints to images scanned.
	// If 1.6 had images "gloo", "discovery" and 1.7 introduced a new image "rate-limit",
	// the map would look like:
	/*
	   ' >= 1.6': ["gloo", "discovery"]
	   ' >= 1.7': ["gloo", "discovery", "rate-limit"]
	*/
	// where the patch number is explicitly not set so that these versions can match all
	// 1.6.x-x releases
	ImagesPerVersion map[string][]string
	// VersionConstraint on releases to security scan
	// any releases that do not pass this constraint will not be security scanned.
	// If left empty, all versions will be scanned
	VersionConstraint *semver.Constraints

	// Required: image repo (quay.io, grc.io, gchr.io)
	ImageRepo string

	// Uploads Sarif file to github security code-scanning results
	// e.g. https://github.com/solo-io/gloo/security/code-scanning
	UploadCodeScanToGithub bool

	// Creates github issue if image vulnerabilities are found
	CreateGithubIssuePerVersion bool

	// Only create github issue if:
	// 	1. Image vulnerabilities are found
	//	2. The version is the latest patch version (Major.Minor.Patch)
	// If set to true, will override the behavior of CreateGithubIssuePerVersion
	CreateGithubIssueForLatestPatchVersion bool

	// Additional context to add to the top of the generated vulnerability report.
	// Example: This could be used to provide debug instructions to developers.
	AdditionalContext string
}

// Main method to call on SecurityScanner which generates .md and .sarif files
// in OutputDir as defined above per repo. If UploadCodeScanToGithub is true,
// sarif files will be uploaded to the repository's code-scanning endpoint.
func (s *SecurityScanner) GenerateSecurityScans(ctx context.Context) error {
	logger := contextutils.LoggerFrom(ctx)

	var err error
	s.githubClient, err = githubutils.GetClient(ctx)
	if err != nil {
		return eris.Wrap(err, "error initializing github client")
	}
	markdownTplFile, err := GetTemplateFile(MarkdownTrivyTemplate)
	if err != nil {
		return eris.Wrap(err, "error creating temporary markdown template file to pass to trivy")
	}
	sarifTplFile, err := GetTemplateFile(SarifTrivyTemplate)
	if err != nil {
		return eris.Wrap(err, "error creating temporary markdown template file to pass to trivy")
	}
	defer func() {
		os.Remove(markdownTplFile)
		os.Remove(sarifTplFile)
	}()

	for _, repo := range s.Repos {
		// Process the user defined options, and configure the non-user controller properties of a SecurityScanRepo
		err := s.initializeRepoConfiguration(ctx, repo)
		if err != nil {
			return err
		}

		for _, release := range repo.releasesToScan {
			releaseStart := time.Now()
			err = repo.RunMarkdownScan(ctx, release, markdownTplFile)
			if err != nil {
				return eris.Wrapf(err, "error generating markdown file from security scan for version %s", release.GetTagName())
			}

			// Only generate sarif files if we are uploading code scan results to github
			if repo.Opts.UploadCodeScanToGithub {
				err = repo.runGithubSarifScan(ctx, release, sarifTplFile)
				if err != nil {
					return eris.Wrapf(err, "error generating github sarif file from security scan for version %s", release.GetTagName())
				}
			}
			logger.Debugf("Completed running markdown scan for release %s after %s", release.GetTagName(), time.Since(releaseStart).String())
		}

	}
	return nil
}

// initializeRepoConfiguration processes the user defined options
// and configures the non-user controller properties of a SecurityScanRepo
func (s *SecurityScanner) initializeRepoConfiguration(ctx context.Context, repo *SecurityScanRepo) error {
	logger := contextutils.LoggerFrom(ctx)
	logger.Debugf("Processing user defined configuration for repository (%s, %s)", repo.Owner, repo.Repo)

	// Ensure Trivy is installed and on PATH
	_, err := exec.LookPath("trivy")
	if err != nil {
		return eris.Wrap(err, "trivy is not on PATH, make sure that the trivy is installed and on PATH")
	}

	repoOptions := repo.Opts

	// Set the Predicate used to filter releases we wish to scan
	repo.scanReleasePredicate = NewSecurityScanRepositoryReleasePredicate(repoOptions.VersionConstraint)

	logger.Debugf("Scanning github repo for releases that match version constraint: %s", repoOptions.VersionConstraint)

	// Get the full set of releases that we expect to scan
	maxReleasesToScan := math.MaxInt32
	releasesToScan, err := githubutils.GetRepoReleasesWithPredicateAndMax(ctx, s.githubClient, repo.Owner, repo.Repo, repo.scanReleasePredicate, maxReleasesToScan)
	if err != nil {
		return eris.Wrapf(err, "unable to fetch all github releases for github.com/%s/%s", repo.Owner, repo.Repo)
	}
	githubutils.SortReleasesBySemver(releasesToScan)
	repo.releasesToScan = releasesToScan

	logger.Debugf("Number of github releases to scan: %d", len(releasesToScan))

	// Initialize a local store of GitHub issues if we will be creating new issues
	githubRepo := GithubRepo{
		RepoName: repo.Repo,
		Owner:    repo.Owner,
	}
	// Default to not creating any issues
	var issuePredicate githubutils.RepositoryReleasePredicate = &githubutils.NoReleasesPredicate{}
	if repoOptions.CreateGithubIssuePerVersion {
		// Create Github issue for all releases, if configured
		issuePredicate = &githubutils.AllReleasesPredicate{}
	}

	if repoOptions.CreateGithubIssueForLatestPatchVersion {
		// Create Github issues for all releases in the set
		issuePredicate = NewLatestPatchRepositoryReleasePredicate(releasesToScan)
	}
	repo.githubIssueWriter = NewGithubIssueWriter(githubRepo, s.githubClient, issuePredicate)
	logger.Debugf("GithubIssueWriter configured with Predicate: %+v", issuePredicate)

	repo.trivyScanner = NewTrivyScanner(executils.CombinedOutputWithStatus)

	logger.Debugf("Completed processing user defined configuration.")
	return nil
}

func (r *SecurityScanRepo) RunMarkdownScan(ctx context.Context, release *github.RepositoryRelease, markdownTplFile string) error {
	// We can swallow the error here, any releases with improper tag names
	// will not be included in the filtered list
	versionToScan, _ := semver.NewVersion(release.GetTagName())
	images, err := r.GetImagesToScan(versionToScan)
	if err != nil {
		return err
	}
	version := versionToScan.String()
	outputDir := path.Join(r.Opts.OutputDir, r.Repo, "markdown_results", version)
	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return err
	}
	var vulnerabilityMd string
	for _, image := range images {
		var imageWithRepo string
		// if the image contains the repo in it (gcr.io/gloo/image-name), we don't use the Opts.ImageRepo
		if strings.Contains(image, "/") {
			imageWithRepo = fmt.Sprintf("%s:%s", image, version)
		} else {
			imageWithRepo = fmt.Sprintf("%s/%s:%s", r.Opts.ImageRepo, image, version)
		}
		fileName := fmt.Sprintf("%s_cve_report.docgen", image)
		output := path.Join(outputDir, fileName)
		_, vulnFound, err := r.trivyScanner.ScanImage(ctx, imageWithRepo, markdownTplFile, output)
		if err != nil {
			if !errors.Is(err, ImageNotFoundError) {
				return eris.Wrapf(err, "error running image scan on image %s", imageWithRepo)
			}
			vulnerabilityMd += fmt.Sprintf("# %s\n\n %s", imageWithRepo, ImageNotFoundError)
		}

		if vulnFound {
			trivyScanMd, err := ioutil.ReadFile(output)
			if err != nil {
				return eris.Wrapf(err, "error reading trivy markdown scan file %s to generate github issue", output)
			}
			vulnerabilityMd += fmt.Sprintf("# %s\n\n %s\n\n", imageWithRepo, trivyScanMd)
		}

	}
	// Create / Update Github issue for the repo if a vulnerability is found
	return r.githubIssueWriter.CreateUpdateVulnerabilityIssue(ctx, release, vulnerabilityMd, r.Opts.AdditionalContext)
}

func (r *SecurityScanRepo) runGithubSarifScan(ctx context.Context, release *github.RepositoryRelease, sarifTplFile string) error {
	// We can swallow the error here, any releases with improper tag names
	// will not be included in the filtered list
	versionToScan, _ := semver.NewVersion(release.GetTagName())

	images, err := r.GetImagesToScan(versionToScan)
	if err != nil {
		return err
	}
	version := versionToScan.String()
	outputDir := path.Join(r.Opts.OutputDir, r.Repo, "sarif_results", version)
	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return err
	}
	for _, image := range images {
		var imageWithRepo string
		// if the image contains the repo in it (gcr.io/gloo/image-name), we don't use the Opts.ImageRepo
		if strings.Contains(image, "/") {
			imageWithRepo = fmt.Sprintf("%s:%s", image, version)
		} else {
			imageWithRepo = fmt.Sprintf("%s/%s:%s", r.Opts.ImageRepo, image, version)
		}
		fileName := fmt.Sprintf("%s_cve_report.sarif", image)
		output := path.Join(outputDir, fileName)
		success, _, err := r.trivyScanner.ScanImage(ctx, imageWithRepo, sarifTplFile, output)
		if err != nil {
			return eris.Wrapf(err, "error running image scan on image %s", imageWithRepo)
		}
		if success {
			fmt.Printf("Security scan for version %s uploaded to github repo github.com/%s/%s\n", version, r.Owner, r.Repo)
			err = r.UploadSecurityScanToGithub(output, version)
			if err != nil {
				return eris.Wrapf(err, "error uploading security scan results sarif to github for version %s", version)
			}
		}
	}
	return nil
}

func (r *SecurityScanRepo) GetImagesToScan(versionToScan *semver.Version) ([]string, error) {
	imagesToScan := map[string]interface{}{}
	for constraintString, images := range r.Opts.ImagesPerVersion {
		constraint, err := semver.NewConstraint(constraintString)
		if err != nil {
			return nil, eris.Wrapf(err, "Error with constraint %s", constraint)
		}
		if constraint.Check(versionToScan) {
			// For each constraint that the current version to scan passes, we add those images to
			// the set of images to scan
			for _, i := range images {
				imagesToScan[i] = true
			}
		}

	}
	if len(imagesToScan) == 0 {
		return nil, eris.Errorf("version %s matched no constraints and has no images to scan", versionToScan.String())
	}
	return stringutils.Keys(imagesToScan), nil
}

type SarifMetadata struct {
	Ref       string `json:"ref"`
	CommitSha string `json:"commit_sha"`
	Sarif     string `json:"sarif"`
}

// Uploads Github security scan in .sarif file format to Github Security Tab under "Code Scanning"
func (r *SecurityScanRepo) UploadSecurityScanToGithub(fileName, versionTag string) error {
	sha, err := githubutils.GetCommitForTag(r.Owner, r.Repo, "v"+versionTag, true)
	if err != nil {
		return err
	}

	githubToken, err := githubutils.GetGithubToken()
	if err != nil {
		return err
	}
	sarifFileBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return eris.Wrapf(err, "error reading sarif file %s", fileName)
	}
	var sarifFile bytes.Buffer
	gzipWriter := gzip.NewWriter(&sarifFile)
	_, err = gzipWriter.Write(sarifFileBytes)
	if err != nil {
		return eris.Wrap(err, "error writing gzip file")
	}
	gzipWriter.Close()
	if len(sha) != 40 {
		return eris.Errorf("Invalid SHA (%s) for version %s", sha, versionTag)
	}
	sarifMetadata := SarifMetadata{
		Ref:       fmt.Sprintf("refs/tags/v%s", versionTag),
		CommitSha: sha,
		Sarif:     base64.StdEncoding.EncodeToString(sarifFile.Bytes()),
	}
	header := req.Header{
		"Authorization": fmt.Sprintf("token %s", githubToken),
		"Content-Type":  "application/json",
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/code-scanning/sarifs", r.Owner, r.Repo)
	res, err := req.Post(url, req.BodyJSON(sarifMetadata), header)
	if err != nil || res.Response().StatusCode != 200 {
		return eris.Wrapf(err, "error uploading sarif file to github, response: \n%s", res)
	}
	fmt.Printf("Response from API, uploading sarif %s: \n %s\n", fileName, res.String())
	return nil
}
