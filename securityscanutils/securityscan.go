package securityscanutils

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/solo-io/go-utils/osutils/executils"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/Masterminds/semver/v3"
	"github.com/solo-io/go-utils/stringutils"

	"github.com/google/go-github/v32/github"
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

	// Creates github issue if image vulnerabilities are found
	CreateGithubIssuePerVersion bool

	// Only create github issue if:
	// 	1. Image vulnerabilities are found
	//	2. The version is the latest patch version (Major.Minor.Patch)
	// If set to true, will override the behavior of CreateGithubIssuePerVersion
	CreateGithubIssueForLatestPatchVersion bool
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
	defer func() {
		_ = os.Remove(markdownTplFile)
	}()

	for _, repo := range s.Repos {
		// Process the user defined options, and configure the non-user controlled properties of a SecurityScanRepo
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
	return r.githubIssueWriter.CreateUpdateVulnerabilityIssue(ctx, release, vulnerabilityMd)
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