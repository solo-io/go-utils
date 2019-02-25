package changelogutils

import (
	"context"
	"github.com/solo-io/go-utils/githubutils"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

type ChangelogEntryType int

const (
	BREAKING_CHANGE ChangelogEntryType = iota
	FIX
	NEW_FEATURE
	NON_USER_FACING
)

type ChangelogEntry struct {
	Type        ChangelogEntryType
	Description string
}

type Changelog struct {
	Entries []ChangelogEntry
	Summary string
	Version string
}

const (
	ChangelogDirectory = "changelog"
)

// Should return the last released version
// Executes git commands, so this won't work if running from an unzipped archive of the code.
func getLatestTag(ctx context.Context, owner, repo string) (string, error) {
	client, err := githubutils.GetClient(ctx)
	if err != nil {
		return "", err
	}
	return githubutils.FindLatestReleaseTag(ctx, client, owner, repo)
}

// Should return the next version to release, based on the names of the subdirectories in the changelog
// Will return an error if there is no version, or multiple versions, larger than the latest tag,
// according to semver
func getProposedTag(ctx context.Context, owner, repo string) (string, error) {
	latestTag, err := getLatestTag(ctx, owner, repo)
	if err != nil {
		return "", err
	}
	subDirs, err := ioutil.ReadDir(ChangelogDirectory)
	if err != nil {
		return "", errors.Wrapf(err, "Error reading changelog directory")
	}
	proposedVersion := ""
	for _, subDir := range subDirs {
		if !subDir.IsDir() {
			return "", errors.Errorf("Unexpected entry %s in changelog directory", subDir.Name())
		}
		if !isValidTag(subDir.Name()) {
			return "", errors.Errorf("Directory name %s is not valid, must be of the form 'vX.Y.Z'", subDir.Name())
		}
		greaterThan, err := isGreaterThan(latestTag, subDir.Name())
		if err != nil {
			return "", err
		}
		if greaterThan {
			if proposedVersion != "" {
				return "", errors.Errorf("Versions %s and %s are both greater than latest tag", subDir.Name(), proposedVersion)
			}
			proposedVersion = subDir.Name()
		}
	}
	err = nil
	if proposedVersion == "" {
		err = errors.Errorf("No version greater than %s found", latestTag)
	}
	return proposedVersion, err
}

func isGreaterThan(greaterTag, lesserTag string) (bool, error) {
	greaterVersion, err := parseVersion(greaterTag)
	if err != nil {
		return false, err
	}
	lesserVersion, err := parseVersion(lesserTag)
	if err != nil {
		return false, err
	}

	if greaterVersion.Major > lesserVersion.Major {
		return true, nil
	} else if greaterVersion.Major < lesserVersion.Major {
		return false, nil
	}

	if greaterVersion.Minor > lesserVersion.Minor {
		return true, nil
	} else if greaterVersion.Minor < lesserVersion.Minor {
		return false, nil
	}

	if greaterVersion.Patch > lesserVersion.Patch {
		return true, nil
	} else if greaterVersion.Patch < lesserVersion.Patch {
		return false, nil
	}

	return false, nil
}

type Version struct {
	Major int
	Minor int
	Patch int
}

func parseVersion(tag string) (*Version, error) {
	if !strings.HasPrefix(tag, "v") {
		return nil, errors.Errorf("Tag %s is not a valid version", tag)
	}
	versionString := tag[1:]
	versionParts := strings.Split(versionString, ".")
	if len(versionParts) != 3 {
		return nil, errors.Errorf("Version %s is not a valid semver version", versionString)
	}
	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return nil, errors.Errorf("Major version %s is not valid", versionParts[0])
	}
	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return nil, errors.Errorf("Minor version %s is not valid", versionParts[1])
	}
	patch, err := strconv.Atoi(versionParts[2])
	if err != nil {
		return nil, errors.Errorf("Patch version %s is not valid", versionParts[2])
	}

	version := Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
	return &version, nil
}

func isValidTag(tag string) bool {
	regex := regexp.MustCompile("(v[0-9]+[.][0-9]+[.][0-9]+$)")
	return regex.MatchString(tag)
}

func readChangelogFile(path string) (*Changelog, error) {
	var changelog Changelog
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed reading changelog file: %s", path)
	}

	if err := yaml.Unmarshal(bytes, changelog); err != nil {
		return nil, errors.Wrap(err, "failed parsing changelog file")
	}

	return &changelog, nil
}

