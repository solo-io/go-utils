package changelogutils

import (
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"io/ioutil"
	"path/filepath"
)

// Should return the next version to release, based on the names of the subdirectories in the changelog
// Will return an error if there is no version, or multiple versions, larger than the latest tag,
// according to semver
func GetProposedTagLocal(latestTag, changelogParentPath string) (string, error) {
	changelogPath := filepath.Join(changelogParentPath, ChangelogDirectory)
	subDirs, err := ioutil.ReadDir(changelogPath)
	if err != nil {
		return "", errors.Wrapf(err, "Error reading changelog directory")
	}
	proposedVersion := ""
	for _, subDir := range subDirs {
		if !subDir.IsDir() {
			return "", errors.Errorf("Unexpected entry %s in changelog directory", subDir.Name())
		}
		proposedVersion, err = validVersion(subDir.Name(), proposedVersion, latestTag)
		if err != nil {
			return "", err
		}
	}
	err = nil
	if proposedVersion == "" {
		err = errors.Errorf("No version greater than %s found", latestTag)
	}
	return proposedVersion, err
}

func ReadChangelogFile(path string) (*Changelog, error) {
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
