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
	"strconv"
	"strings"
	"time"

	"github.com/solo-io/go-utils/stringutils"
	"github.com/solo-io/go-utils/versionutils"

	"github.com/Masterminds/semver/v3"

	"github.com/google/go-github/v32/github"
	"github.com/solo-io/go-utils/osutils/executils"

	"github.com/imroc/req"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/log"
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

	// The RepositoryReleasePredicate used to determine if a particular release
	// should be run through our scanner
	scanReleasePredicate githubutils.RepositoryReleasePredicate

	// The RepositoryReleasePredicate used to determine if a vulnerability
	// associated with a certain release should be uploaded to GitHub
	createGithubIssuePredicate githubutils.RepositoryReleasePredicate

	// A local cache of all existing GitHub issues
	// Used to ensure that we are updating existing issues that were created
	// by previous scans
	allGithubIssues []*github.Issue
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
}

// Status code returned by Trivy if a vulnerability is found
const VulnerabilityFoundStatusCode = 52

// Labels that are applied to github issues that security scan generates
var TrivyLabels = []string{"trivy", "vulnerability"}

// Main method to call on SecurityScanner which generates .md and .sarif files
// in OutputDir as defined above per repo. If UploadCodeScanToGithub is true,
// sarif files will be uploaded to the repository's code-scanning endpoint.
func (s *SecurityScanner) GenerateSecurityScans(ctx context.Context) error {
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

		maxReleasesToScan := math.MaxInt32
		partialFilteredReleases, err := githubutils.GetRepoReleasesWithPredicateAndMax(ctx, s.githubClient, repo.Owner, repo.Repo, repo.scanReleasePredicate, maxReleasesToScan)
		if err != nil {
			return eris.Wrapf(err, "unable to fetch all github releases for github.com/%s/%s", repo.Owner, repo.Repo)
		}
		githubutils.SortReleasesBySemver(partialFilteredReleases)
		filteredReleases := []*github.RepositoryRelease{}

		// We could use maxint but we dont really care
		// as we can just check if major minor changed
		recentMajor := -1
		recentMinor := -1
		for _, release := range partialFilteredReleases {
			version, err := versionutils.ParseVersion(release.GetTagName())
			if err != nil {
				continue
			}
			if version.Major == recentMajor && version.Minor == recentMinor {
				continue
			}

			// This is the largest patch release
			recentMajor = version.Major
			recentMinor = version.Minor
			filteredReleases = append(filteredReleases, release)
		}

		for _, release := range filteredReleases {
			err = repo.RunMarkdownScan(ctx, s.githubClient, release, markdownTplFile)
			if err != nil {
				return eris.Wrapf(err, "error generating markdown file from security scan for version %s", release.GetTagName())
			}
			// Only generate sarif files if we are uploading code scan results
			// to github
			if repo.Opts.UploadCodeScanToGithub {
				err = repo.RunGithubSarifScan(release, sarifTplFile)
				if err != nil {
					return eris.Wrapf(err, "error generating github sarif file from security scan for version %s", release.GetTagName())
				}
			}
		}

	}
	return nil
}

func (s *SecurityScanner) initializeRepoConfiguration(ctx context.Context, repo *SecurityScanRepo) error {
	repoOptions := repo.Opts

	// Initialize a local store of GitHub issues if we will be creating new issues
	if repoOptions.CreateGithubIssuePerVersion {
		issues, err := githubutils.GetAllIssues(ctx, s.githubClient, repo.Owner, repo.Repo, &github.IssueListByRepoOptions{
			State:  "open",
			Labels: TrivyLabels,
		})
		if err != nil {
			return eris.Wrapf(err, "error fetching all issues from github.com/%s/%s", repo.Owner, repo.Repo)
		}
		repo.allGithubIssues = issues
	}

	// Set the Predicate used to filter releases we wish to scan
	repo.scanReleasePredicate = &SecurityScanRepositoryReleasePredicate{
		versionConstraint: repoOptions.VersionConstraint,
	}

	// Default to not creating any issues
	repo.createGithubIssuePredicate = &githubutils.NoReleasesPredicate{}
	if repoOptions.CreateGithubIssuePerVersion {
		// Create Github issue for all releases, if configured
		repo.createGithubIssuePredicate = &githubutils.AllReleasesPredicate{}
	}

	// TODO Add logic to handle instantiating a Predicate that returns true only if Release matches latest LTS

	return nil
}

func (r *SecurityScanRepo) RunMarkdownScan(ctx context.Context, client *github.Client, release *github.RepositoryRelease, markdownTplFile string) error {
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
		_, vulnFound, err := RunTrivyScan(imageWithRepo, version, markdownTplFile, output)
		if err != nil {
			return eris.Wrapf(err, "error running image scan on image %s", imageWithRepo)
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
	// and CreateGithubIssuePerVersion is set to true
	if r.createGithubIssuePredicate.Apply(release) {
		if vulnerabilityMd == "" {
			// We did not find vulnerabilities in any of the images for this version
			// do not create an empty github issue
			return nil
		}
		return r.CreateUpdateVulnerabilityIssue(ctx, client, version, vulnerabilityMd)
	}
	return nil
}

func (r *SecurityScanRepo) RunGithubSarifScan(release *github.RepositoryRelease, sarifTplFile string) error {
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
		success, _, err := RunTrivyScan(imageWithRepo, version, sarifTplFile, output)
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

// The SecurityScanRepositoryReleasePredicate is responsible for defining which
// github.RepositoryRelease artifacts should be included in the bulk security scan
// At the moment, the two requirements are that:
// 1. The release is not a pre-release or draft
// 2. The release matches the configured version constraint
type SecurityScanRepositoryReleasePredicate struct {
	versionConstraint *semver.Constraints
}

func (s *SecurityScanRepositoryReleasePredicate) Apply(release *github.RepositoryRelease) bool {
	if release.GetPrerelease() || release.GetDraft() {
		// Do not include pre-releases or drafts
		return false
	}

	versionToTest, err := semver.NewVersion(release.GetTagName())
	if err != nil {
		// If we are unable to parse the release version, we do not include it in the filtered list
		log.Warnf("unable to parse release version %s", release.GetTagName())
		return false
	}

	if !s.versionConstraint.Check(versionToTest) {
		// If the release version does not pass the version constraint, do not include
		return false
	}

	// If all checks have passed, include the release
	return true
}

// Runs trivy scan command
// Returns (trivy scan ran successfully, vulnerabilities found, error running trivy scan)
func RunTrivyScan(image, version, templateFile, output string) (bool, bool, error) {
	// Ensure Trivy is installed and on PATH
	_, err := exec.LookPath("trivy")
	if err != nil {
		return false, false, eris.Wrap(err, "trivy is not on PATH, make sure that the trivy v0.18 is installed and on PATH")
	}
	trivyScanArgs := []string{"image",
		// Trivy will return a specific status code (which we have specified) if a vulnerability is found
		"--exit-code", strconv.Itoa(VulnerabilityFoundStatusCode),
		"--severity", "HIGH,CRITICAL",
		"--format", "template",
		"--template", "@" + templateFile,
		"--output", output,
		image}
	// Execute the trivy scan, with retries and sleep's between each retry
	// This can occur due to connectivity issues or epehemeral issues with
	// the registery. For example sometimes quay has issues providing a given layer
	// This leads to a total wait time of up to 110 seconds outside of the base
	// operation. This timing is in the same ballpark as what k8s finds sensible
	out, statusCode, err := executeTrivyScanWithRetries(
		trivyScanArgs, 5,
		func(attempt int) { time.Sleep(time.Duration((attempt^2)*2) * time.Second) },
	)

	// Check if a vulnerability has been found
	vulnFound := statusCode == VulnerabilityFoundStatusCode
	// err will be non-nil if there is a non-zero status code
	// so if the status code is the special "vulnerability found" status code,
	// we don't want to report it as a regular error
	if !vulnFound && err != nil {
		// delete empty trivy output file that may have been created
		_ = os.Remove(output)
		// swallow error if image is not found error, so that we can continue scanning releases
		// even if some releases failed and we didn't publish images for those releases
		// this error used to happen if a release was a pre-release and therefore images
		// weren't pushed to the container registry.
		// we have since filtered out non-release images from being scanned so this warning
		// shouldn't occur, but leaving here in case there was another edge case we missed
		if IsImageNotFoundErr(string(out)) {
			log.Warnf("image %s not found for version %s", image, version)
			return false, false, nil
		}
		return false, false, eris.Wrapf(err, "error running trivy scan on image %s, version %s, Logs: \n%s", image, version, string(out))
	}
	return true, vulnFound, nil
}

func executeTrivyScanWithRetries(trivyScanArgs []string, retryCount int,
	backoffStrategy func(int)) ([]byte, int, error) {
	if retryCount == 0 {
		retryCount = 5
	}
	if backoffStrategy == nil {
		backoffStrategy = func(attempt int) {
			time.Sleep(time.Second)
		}
	}

	var (
		out        []byte
		statusCode int
		err        error
	)

	for attempt := 0; attempt < retryCount; attempt++ {
		trivyScanCmd := exec.Command("trivy", trivyScanArgs...)
		out, statusCode, err = executils.CombinedOutputWithStatus(trivyScanCmd)

		// If there is no error, don't retry
		if err == nil {
			return out, statusCode, err
		}

		// If there is no image, don't retry
		if IsImageNotFoundErr(string(out)) {
			return out, statusCode, err
		}

		backoffStrategy(attempt)
	}
	return out, statusCode, err
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

// Creates/Updates a Github Issue per image
// The github issue will have the markdown table report of the image's vulnerabilities
// example: https://github.com/solo-io/solo-projects/issues/2458
func (r *SecurityScanRepo) CreateUpdateVulnerabilityIssue(ctx context.Context, client *github.Client, version, vulnerabilityMarkdown string) error {
	issueTitle := fmt.Sprintf("Security Alert: %s", version)
	issueRequest := &github.IssueRequest{
		Title:  github.String(issueTitle),
		Body:   github.String(vulnerabilityMarkdown),
		Labels: &TrivyLabels,
	}
	createNewIssue := true

	for _, issue := range r.allGithubIssues {
		// If issue already exists, update existing issue with new security scan
		if issue.GetTitle() == issueTitle {
			// Only create new issue if issue does not already exist
			createNewIssue = false
			err := githubutils.UpdateIssue(ctx, client, r.Owner, r.Repo, issue.GetNumber(), issueRequest)
			if err != nil {
				return eris.Wrapf(err, "error updating issue with issue request %+v", issueRequest)
			}
			break
		}
	}
	if createNewIssue {
		_, err := githubutils.CreateIssue(ctx, client, r.Owner, r.Repo, issueRequest)
		if err != nil {
			return eris.Wrapf(err, "error creating issue with issue request %+v", issueRequest)
		}
	}
	return nil
}

func IsImageNotFoundErr(logs string) bool {
	if strings.Contains(logs, "No such image: ") {
		return true
	}
	return false
}
