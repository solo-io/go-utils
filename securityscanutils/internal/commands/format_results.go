package commands

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v32/github"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/securityscanutils"
	"github.com/solo-io/go-utils/securityscanutils/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func FormatResultsCommand(ctx context.Context, globalFlags *internal.GlobalFlags) *cobra.Command {
	opts := &formatResultsOptions{
		GlobalFlags: globalFlags,
	}

	cmd := &cobra.Command{
		Use:     "format-results",
		Aliases: []string{"gen-security-scan-md"}, // this is the previous name of the command
		Short:   "[TODO]",
		RunE: func(cmd *cobra.Command, args []string) error {
			return doFormatResults(ctx, opts)
		},
	}
	opts.addToFlags(cmd.Flags())

	return cmd
}

type formatResultsOptions struct {
	*internal.GlobalFlags

	// Values that are empty/nil if unset
	targetRepo             string
	targetRepoWritten      string
	repoCachedReleasesFile string
	minScannedVersion      string
	imageFile              string
	// Values with defaults
	repoOwner              string
	imageRepo              string
	generateCachedReleases bool
	// Derived values that aren't directly related to inputs
	allImages []string
}

func (f *formatResultsOptions) addToFlags(flags *pflag.FlagSet) {

	// add args/flags
	flags.StringVarP(&f.targetRepo, "TargetRepo", "r", "", "The repository to scan")
	flags.StringVarP(&f.targetRepoWritten, "TargetRepoWritten", "w", "",
		"Specify the human readable name of the repository to scan for output purposes.")
	flags.StringVarP(&f.repoCachedReleasesFile, "CachedReleasesFile", "c", "",
		"The name of the file that contains a list of all releases from the given repository."+
			" This file is generated by the 'gen-releases' command, and used by the others.")
	flags.StringVarP(&f.minScannedVersion, "MinScannedVersion", "m", "",
		"The minimum version of images to scan. If set, will scan every image from this to the present, and will scan all images otherwise")
	flags.BoolVarP(&f.generateCachedReleases, "GenerateCachedReleases", "p", true,
		"If true, then populate the file specified by the CachedReleasesFile flag with all releases from Github."+
			" If false, then the command assumes that the file has already been created and populated. "+
			" Should be set to false for testing to avoid rate-limiting by Github. Defaults to true.")
	flags.StringVarP(&f.imageFile, "ImageFile", "f", "",
		"Different release versions may have different images to scan."+
			"\nTo deal with this, the run-security-scan command expects a file input that maps version constraints to images"+
			"\nto be scanned if a version matches that constraint. Constraints must be mutually exclusive."+
			"\nThe file is expected to be a csv, where the first element of each line is the constraint, and every subsequent element"+
			"\nin that line is an image to be scanned if that constraint is matched."+
			"\nRead https://github.com/Masterminds/semver#checking-version-constraints for more about how to use semver constraints.")
	flags.StringVarP(&f.repoOwner, "RepoOwner", "", securityscanutils.GithubRepositoryOwner,
		"The owner of the repository to scan. Defaults to 'solo-io'")
	flags.StringVarP(&f.imageRepo, "ImageRepo", "", "quay.io/solo-io",
		"The repository where images to scan are located. Defaults to 'quay.io/solo-io'")

	// mark required args
	cliutils.MustMarkFlagRequired(flags, "TargetRepo")
	cliutils.MustMarkFlagRequired(flags, "TargetRepoWritten")
	cliutils.MustMarkFlagRequired(flags, "ImageFile")

}

// copied verbatim from: https://github.com/solo-io/go-utils/blob/877b67c2d5c4eee8bf710bd4c0337188698396dc/securityscanutils/security_scan_command.go#L266
func doFormatResults(ctx context.Context, opts *formatResultsOptions) error {
	// Initialize Auth
	client, err := githubutils.GetClient(ctx)
	// Sets the opts.allImages value, which we need for this command
	readImageVersionConstraintsFile(opts)
	if err != nil {
		return err
	}
	var allReleases []*github.RepositoryRelease
	if len(opts.repoCachedReleasesFile) == 0 {
		allReleases = getCachedReleases(opts.repoCachedReleasesFile)
	} else {
		allReleases, err = githubutils.GetAllRepoReleases(ctx, client, securityscanutils.GithubRepositoryOwner, opts.targetRepo)
		if err != nil {
			return err
		}
	}
	githubutils.SortReleasesBySemver(allReleases)
	versionsToScan := getVersionsToScan(opts, allReleases)
	return BuildSecurityScanReportForRepo(versionsToScan, opts)
}

func getCachedReleases(fileName string) []*github.RepositoryRelease {
	bArray, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil
	}
	buf := bytes.NewBuffer(bArray)
	enc := gob.NewDecoder(buf)
	var releases []*github.RepositoryRelease
	err = enc.Decode(&releases)
	if err != nil {
		return nil
	}
	return releases
}

func getVersionsToScan(opts *formatResultsOptions, releases []*github.RepositoryRelease) []string {
	var (
		versions             []string
		stableOnlyConstraint *semver.Constraints
		err                  error
	)
	minVersionToScan := opts.minScannedVersion
	if minVersionToScan == "" {
		log.Println("MinScannedVersion flag not set, scanning all versions from repo")
	} else {
		stableOnlyConstraint, err = semver.NewConstraint(fmt.Sprintf(">= %s", minVersionToScan))
		if err != nil {
			log.Fatalf("Invalid constraint version: %s", minVersionToScan)
		}
	}

	for _, release := range releases {
		// ignore beta releases when display security scan results
		test, err := semver.NewVersion(release.GetTagName())
		if err != nil {
			continue
		}
		if stableOnlyConstraint == nil || stableOnlyConstraint.Check(test) {
			versions = append(versions, test.String())
		}
	}
	return versions
}

func BuildSecurityScanReportForRepo(tags []string, opts *formatResultsOptions) error {
	// tags are sorted by minor version
	latestTag := tags[0]
	prevMinorVersion, _ := semver.NewVersion(latestTag)
	for ix, tag := range tags {
		semver, err := semver.NewVersion(tag)
		if err != nil {
			return err
		}
		if ix == 0 || semver.Minor() != prevMinorVersion.Minor() {
			fmt.Printf("\n***Latest %d.%d.x %s Release: %s***\n\n", semver.Major(), semver.Minor(), opts.targetRepoWritten, tag)
			err = printImageReportForRepo(tag, opts)
			if err != nil {
				return err
			}
			prevMinorVersion = semver
		} else {
			fmt.Printf("<details><summary> Release %s </summary>\n\n", tag)
			err = printImageReportForRepo(tag, opts)
			if err != nil {
				return err
			}
			fmt.Println("</details>")
		}
	}

	return nil
}

func printImageReportForRepo(tag string, opts *formatResultsOptions) error {
	for _, image := range opts.allImages {
		fmt.Printf("**%s %s image**\n\n", opts.targetRepoWritten, image)
		url := "https://storage.googleapis.com/solo-gloo-security-scans/" + opts.targetRepo + "/" + tag + "/" + image + "_cve_report.docgen"
		report, err := GetSecurityScanReport(url)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n\n", report)
	}
	return nil
}

func GetSecurityScanReport(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	var report string
	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		report = string(bodyBytes)
	} else if resp.StatusCode == http.StatusNotFound {
		// Older releases may be missing scan results
		report = "No scan found\n"
	}
	resp.Body.Close()

	return report, nil
}

// Reads in a file, and tries to turn it into a map from version constraints to lists of images
// As a byproduct, it also caches all unique images found into the option field 'allImages'
// Also, I'm not sure why I didn't just use a csv reader... oh well.
func readImageVersionConstraintsFile(opts *formatResultsOptions) (map[string][]string, error) {
	imagesPerVersion := make(map[string][]string)
	imageSet := make(map[string]interface{})

	dat, err := ioutil.ReadFile(opts.imageFile)
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(dat), "\n") {
		trimmedLine := strings.TrimSpace(line)
		if len(trimmedLine) == 0 || string(trimmedLine[0]) == "#" {
			continue
		}
		values := strings.Split(trimmedLine, ",")
		if len(values) < 2 {
			return nil, internal.MalformedVersionImageConstraintLine(line)
		}
		for i, _ := range values {
			trimVal := strings.TrimSpace(values[i])
			values[i] = trimVal
			if i > 0 {
				imageSet[trimVal] = nil
			}
		}
		imagesPerVersion[values[0]] = values[1:]
	}
	var allImages []string
	for image, _ := range imageSet {
		allImages = append(allImages, image)
	}
	opts.allImages = allImages

	return imagesPerVersion, nil
}
