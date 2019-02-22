package changelogutils

import (
	"io/ioutil"

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
	directory = "changelog"
)

// Should return the last released version
// Executes git commands, so this won't work if running from an unzipped archive of the code.
func getLatestTag() (string, error) {
	return "", nil
}

// Should return the next version to release, based on the names of the subdirectories in the changelog
// Will return an error if there is no version, or multiple versions, larger than the latest tag,
// according to semver
func getProposedTag() (string, error) {
	return "", nil
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

