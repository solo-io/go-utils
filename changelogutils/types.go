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

