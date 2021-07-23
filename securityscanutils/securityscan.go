package securityscanutils

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/solo-io/go-utils/osutils/executils"

	"github.com/Masterminds/semver/v3"
	"github.com/imroc/req"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/go-utils/osutils"
)

type SecurityScanner struct {
	Repos        []*SecurityScanRepo
	githubClient *github.Client
}

type SecurityScanRepo struct {
	Repo  string
	Owner string
	Opts  *SecurityScanOpts

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
	CreateGithubIssuePerImageVulnerability bool
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
		opts := repo.Opts
		allReleases, err := githubutils.GetAllRepoReleases(ctx, s.githubClient, repo.Owner, repo.Repo)
		if err != nil {
			return eris.Wrapf(err, "unable to fetch all github releases for github.com/%s/%s", repo.Owner, repo.Repo)
		}
		// Filter releases by version constraint provided
		filteredReleases := githubutils.FilterReleases(allReleases, opts.VersionConstraint)
		githubutils.SortReleasesBySemver(filteredReleases)
		if repo.Opts.CreateGithubIssuePerImageVulnerability {
			repo.allGithubIssues, err = githubutils.GetAllIssues(ctx, s.githubClient, repo.Owner, repo.Repo, &github.IssueListByRepoOptions{
				State:  "open",
				Labels: TrivyLabels,
			})
			if err != nil {
				return eris.Wrapf(err, "error fetching all issues from github.com/%s/%s", repo.Owner, repo.Repo)
			}
		}
		for _, release := range filteredReleases {
			// We can swallow the error here, any releases with improper tag names
			// will not be included in the filtered list
			ver, _ := semver.NewVersion(release.GetTagName())
			err = repo.RunMarkdownScan(ctx, s.githubClient, ver, markdownTplFile)
			if err != nil {
				return eris.Wrapf(err, "error generating markdown file from security scan for version %s", release.GetTagName())
			}
			// Only generate sarif files if we are uploading code scan results
			// to github
			if repo.Opts.UploadCodeScanToGithub {
				err = repo.RunGithubSarifScan(ver, sarifTplFile)
				if err != nil {
					return eris.Wrapf(err, "error generating github sarif file from security scan for version %s", release.GetTagName())
				}
			}
		}

	}
	return nil
}

func (r *SecurityScanRepo) RunMarkdownScan(ctx context.Context, client *github.Client, versionToScan *semver.Version, markdownTplFile string) error {
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
	for _, image := range images {
		imageWithRepo := fmt.Sprintf("%s/%s:%s", r.Opts.ImageRepo, image, version)
		fileName := fmt.Sprintf("%s_cve_report.docgen", image)
		output := path.Join(outputDir, fileName)
		_, vulnFound, err := RunTrivyScan(imageWithRepo, version, markdownTplFile, output)
		if err != nil {
			return eris.Wrapf(err, "error running image scan on image %s", imageWithRepo)
		}
		// Create / Update Github issue for the repo if a vulnerability is found
		// and CreateGithubIssuePerImageVulnerability is set to true
		if vulnFound && r.Opts.CreateGithubIssuePerImageVulnerability {
			err = r.CreateUpdateVulnerabilityIssue(ctx, client, imageWithRepo, output)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *SecurityScanRepo) RunGithubSarifScan(versionToScan *semver.Version, sarifTplFile string) error {
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
		imageWithRepo := fmt.Sprintf("%s/%s:%s", r.Opts.ImageRepo, image, version)
		fileName := fmt.Sprintf("%s_cve_report.sarif", image)
		output := path.Join(outputDir, fileName)
		success, _, err := RunTrivyScan(imageWithRepo, version, sarifTplFile, output)
		if err != nil {
			return eris.Wrapf(err, "error running image scan on image %s", imageWithRepo)
		}
		if success {
			err = r.UploadSecurityScanToGithub(output, version)
			if err != nil {
				return eris.Wrapf(err, "error uploading security scan results sarif to github for version %s", version)
			}
		}
	}
	return nil
}

func (r *SecurityScanRepo) GetImagesToScan(versionToScan *semver.Version) ([]string, error) {
	var imagesToScan []string
	for constraintString, images := range r.Opts.ImagesPerVersion {
		constraint, err := semver.NewConstraint(constraintString)
		if err != nil {
			return nil, eris.Wrapf(err, "Error with constraint %s", constraint)
		}
		if constraint.Check(versionToScan) {
			// We want to make sure that each version only matches ONE constraint provided
			// in the constraint -> []images map, so that we are scanning the right images for each version
			if imagesToScan != nil {
				return nil, eris.Errorf(
					"version %s matched more than one constraint provided, please make all constraints mutually exclusive", versionToScan.String())
			}
			imagesToScan = images
		}

	}
	if imagesToScan == nil {
		return nil, eris.Errorf("version %s matched no constraints and has no images to scan", versionToScan.String())
	}
	return imagesToScan, nil
}

// Runs trivy scan command
// Returns (trivy scan ran successfully, vulnerabilities found, error running trivy scan)
func RunTrivyScan(image, version, templateFile, output string) (bool, bool, error) {
	// Ensure Trivy is installed and on PATH
	_, err := exec.LookPath("trivy")
	if err != nil {
		return false, false, eris.Wrap(err, "trivy is not on PATH, make sure that the trivy v0.18 is installed and on PATH")
	}
	args := []string{"image",
		// Trivy will return a specific status code (which we have specified) if a vulnerability is found
		"--exit-code", strconv.Itoa(VulnerabilityFoundStatusCode),
		"--severity", "HIGH,CRITICAL",
		"--format", "template",
		"--template", "@" + templateFile,
		"--output", output,
		image}
	cmd := exec.Command("trivy", args...)
	out, statusCode, err := executils.CombinedOutputWithStatus(cmd)
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
		if IsImageNotFoundErr(string(out)) {
			log.Warnf("image %s not found for version %s", image, version)
			return false, false, nil
		}
		return false, false, eris.Wrapf(err, "error running trivy scan on image %s, version %s, Logs: \n%s", image, version, string(out))
	}
	return true, vulnFound, nil
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
	githubToken, err := osutils.GetEnvE("GITHUB_TOKEN")
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
func (r *SecurityScanRepo) CreateUpdateVulnerabilityIssue(ctx context.Context, client *github.Client, image, markdownScanFilePath string) error {
	issueTitle := fmt.Sprintf("Security Alert: %s", image)
	markdownScan, err := ioutil.ReadFile(markdownScanFilePath)
	if err != nil {
		return eris.Wrapf(err, "error reading file %s", markdownScanFilePath)
	}
	issueRequest := &github.IssueRequest{
		Title:  github.String(issueTitle),
		Body:   github.String(string(markdownScan)),
		Labels: &TrivyLabels,
	}
	createNewIssue := true

	for _, issue := range r.allGithubIssues {
		// If issue already exists, update existing issue with new security scan
		if strings.Contains(issue.GetTitle(), issueTitle) {
			// Only create new issue if issue does not already exist
			createNewIssue = false
			err = githubutils.UpdateIssue(ctx, client, r.Owner, r.Repo, issue.GetNumber(), issueRequest)
			if err != nil {
				return eris.Wrapf(err, "error updating issue with issue request %+v", issueRequest)
			}
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
