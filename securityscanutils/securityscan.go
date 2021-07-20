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
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/imroc/req"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/go-utils/osutils"
)

type SecurityScanner struct {
	Repos []*SecurityScanRepo
}

type SecurityScanRepo struct {
	Repo  string
	Owner string
	Opts  *SecurityScanOpts
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
}

// Main method to call on SecurityScanner which generates .md and .sarif files
// in OutputDir as defined above per repo. If UploadCodeScanToGithub is true,
// sarif files will be uploaded to the repository's code-scanning endpoint.
func (s *SecurityScanner) GenerateSecurityScans(ctx context.Context) error {

	client, err := githubutils.GetClient(ctx)
	if err != nil {
		return eris.Wrap(err, "error initializing github client")
	}
	for _, repo := range s.Repos {
		opts := repo.Opts
		allReleases, err := githubutils.GetAllRepoReleases(ctx, client, repo.Owner, repo.Repo)
		if err != nil {
			return eris.Wrapf(err, "unable to fetch all github releases for github.com/%s/%s", repo.Owner, repo.Repo)
		}
		githubutils.SortReleasesBySemver(allReleases)
		// Filter releases by version constraint provided
		filteredReleases := githubutils.FilterReleases(allReleases, opts.VersionConstraint)
		markdownTplFile, err := GetTemplateFile(MarkdownTrivyTemplate)
		if err != nil {
			return eris.Wrap(err, "error creating temporary markdown template file to pass to trivy")
		}
		sarifTplFile, err := GetTemplateFile(SarifTrivyTemplate)
		if err != nil {
			return eris.Wrap(err, "error creating temporary markdown template file to pass to trivy")
		}
		for _, release := range filteredReleases {
			// We can swallow the error here, any releases with improper tag names
			// will not be included in the filtered list
			ver, _ := semver.NewVersion(release.GetTagName())
			err = repo.RunMarkdownScan(ver, markdownTplFile)
			if err != nil {
				return eris.Wrapf(err, "error generating markdown file from security scan for version %s", release.GetTagName())
			}
			err = repo.RunGithubSarifScan(ver, sarifTplFile)
			if err != nil {
				return eris.Wrapf(err, "error generating github sarif file from security scan for version %s", release.GetTagName())
			}
		}

	}
	return nil
}

func (r *SecurityScanRepo) RunMarkdownScan(versionToScan *semver.Version, markdownTplFile string) error {
	images, err := r.GetImagesToScan(versionToScan)
	if err != nil {
		return err
	}
	version := versionToScan.String()
	outputDir := path.Join(r.Opts.OutputDir, r.Repo, "markdown_results", version)
	err = osutils.CreateDirIfNotExists(outputDir)
	if err != nil {
		return err
	}
	for _, image := range images {
		imageWithRepo := fmt.Sprintf("%s/%s:%s", r.Opts.ImageRepo, image, version)
		fileName := fmt.Sprintf("%s_cve_report.docgen", image)
		output := path.Join(outputDir, fileName)
		err = RunTrivyScan(imageWithRepo, version, markdownTplFile, output)
		if err != nil {
			return eris.Wrapf(err, "error running image scan on image %s", imageWithRepo)
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
	err = osutils.CreateDirIfNotExists(outputDir)
	if err != nil {
		return err
	}
	for _, image := range images {
		imageWithRepo := fmt.Sprintf("%s/%s:%s", r.Opts.ImageRepo, image, version)
		fileName := fmt.Sprintf("%s_cve_report.sarif", image)
		output := path.Join(outputDir, fileName)
		success, err := RunTrivyScan(imageWithRepo, version, sarifTplFile, output)
		if err != nil {
			return eris.Wrapf(err, "error running image scan on image %s", imageWithRepo)
		}
		if success && r.Opts.UploadCodeScanToGithub {
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
				return nil, eris.Errorf("version %s matched more than one constraint provided, please make all constraints"+
					"mutually exclusive", versionToScan.String())
			}
			imagesToScan = images
		}

	}
	return imagesToScan, nil
}

// Runs trivy scan command
// returns if trivy scan ran successfully and error if there was one
func RunTrivyScan(image, version, templateFile, output string) (bool, error) {
	// Ensure Trivy is installed and on PATH
	_, err := exec.LookPath("trivy")
	if err != nil {
		return false, eris.Wrap(err, "trivy is not on PATH, make sure that the trivy v0.18 is installed and on PATH")
	}
	args := []string{"image", "--severity", "HIGH,CRITICAL", "--format", "template", "--template", "@" + templateFile, "--output", output, image}
	cmd := exec.Command("trivy", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// delete empty trivy output file that may have been created
		_ = os.Remove(output)
		// swallow error if image is not found error, so that we can continue scanning releases
		// even if some releases failed and we didn't publish images for those releases
		if IsImageNotFoundErr(string(out)) {
			log.Warnf("image %s not found for version %s", image, version)
			return false, nil
		}
		return false, eris.Wrapf(err, "error running trivy scan on image %s, version %s, Logs: \n%s", image, version, string(out))
	}
	return true, nil
}

type SarifMetadata struct {
	Ref       string `json:"ref"`
	CommitSha string `json:"commit_sha"`
	Sarif     string `json:"sarif"`
}

type Response struct {
	ShaObject `json:"object"`
}

type ShaObject struct {
	Sha string `json:"sha"`
}

func (r *SecurityScanRepo) UploadSecurityScanToGithub(fileName, versionTag string) error {
	githubRepoApiUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s", r.Owner, r.Repo)
	githubToken, err := osutils.GetEnvE("GITHUB_TOKEN")
	if err != nil {
		return err
	}
	resp, err := req.Get(githubRepoApiUrl+"/git/refs/tags/v"+versionTag, req.Header{"Authorization": "token " + githubToken})
	if err != nil {
		return eris.Wrapf(err, "Unable to get commit for version v%s", versionTag)
	}
	shaResp := &Response{}
	resp.ToJSON(shaResp)
	fmt.Printf("%+v\n", shaResp)
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return eris.Wrapf(err, "error reading sarif file %s", fileName)
	}
	var sarifFile bytes.Buffer
	w := gzip.NewWriter(&sarifFile)
	_, err = w.Write(b)
	if err != nil {
		return eris.Wrap(err, "error writing gzip file")
	}
	w.Close()
	if len(shaResp.Sha) != 40 {
		return eris.Errorf("Invalid SHA (%s) for version %s", shaResp.Sha, versionTag)
	}
	sarifMetadata := SarifMetadata{
		Ref:       fmt.Sprintf("refs/tags/v%s", versionTag),
		CommitSha: shaResp.Sha,
		Sarif:     base64.StdEncoding.EncodeToString(sarifFile.Bytes()),
	}
	header := req.Header{
		"Authorization": fmt.Sprintf("token %s", githubToken),
		"Content-Type":  "application/json",
	}
	res, err := req.Post(githubRepoApiUrl+"/code-scanning/sarifs", req.BodyJSON(sarifMetadata), header)
	fmt.Println(res.String())
	if err != nil || res.Response().StatusCode != 200 {
		return eris.Wrapf(err, "error uploading sarif file to github, response: \n%s", res)
	}
	return nil
}

func IsImageNotFoundErr(logs string) bool {
	if strings.Contains(logs, "No such image: ") {
		return true
	}
	return false
}
