package changelogutils

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/versionutils"
	"github.com/spf13/afero"
)

type ChangelogEntry struct {
	Type        ChangelogEntryType
	Description string
	IssueLink   string
}

type ChangelogFile struct {
	Entries []*ChangelogEntry `json:"changelog,omitempty"`
}

type Changelog struct {
	Files   []*ChangelogFile
	Summary string
	Version *versionutils.Version
}

const (
	ChangelogDirectory = "changelog"
	SummaryFile = "summary.md"
)

// Should return the last released version
func GetLatestTag(ctx context.Context, owner, repo string) (string, error) {
	client, err := githubutils.GetClient(ctx)
	if err != nil {
		return "", err
	}
	return githubutils.FindLatestReleaseTag(ctx, client, owner, repo)
}

// Should return the next version to release, based on the names of the subdirectories in the changelog
// Will return an error if there is no version, or multiple versions, larger than the latest tag,
// according to semver
func GetProposedTag(fs afero.Fs, latestTag, changelogParentPath string) (string, error) {
	changelogPath := filepath.Join(changelogParentPath, ChangelogDirectory)
	subDirs, err := afero.ReadDir(fs, changelogPath)
	if err != nil {
		return "", errors.Wrapf(err, "Error reading changelog directory")
	}
	proposedVersion := ""
	for _, subDir := range subDirs {
		if !subDir.IsDir() {
			return "", errors.Errorf("Unexpected entry %s in changelog directory", subDir.Name())
		}
		if !versionutils.MatchesRegex(subDir.Name()) {
			return "", errors.Errorf("Directory name %s is not valid, must be of the form 'vX.Y.Z'", subDir.Name())
		}
		greaterThan, err := versionutils.IsGreaterThanTag(subDir.Name(), latestTag)
		if err != nil {
			return "", err
		}
		if greaterThan {
			if proposedVersion != "" {
				return "", errors.Errorf("Versions %s and %s are both greater than latest tag %s", subDir.Name(), proposedVersion, latestTag)
			}
			proposedVersion = subDir.Name()
		}
	}
	if proposedVersion == "" {
		return "", errors.Errorf("No version greater than %s found", latestTag)
	}
	return proposedVersion, nil
}

func ReadChangelogFile(fs afero.Fs, path string) (*ChangelogFile, error) {
	var changelog ChangelogFile
	bytes, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed reading changelog file: %s", path)
	}

	if err := yaml.Unmarshal(bytes, &changelog); err != nil {
		return nil, errors.Errorf("File %s is not a valid changelog file", path)
	}

	return &changelog, nil
}

func VersionToString(v *versionutils.Version) string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func hasBreakingChange(file *ChangelogFile) bool {
	for _, changelogEntry := range file.Entries {
		if changelogEntry.Type == BREAKING_CHANGE {
			return true
		}
	}
	return false
}

func ComputeChangelog(fs afero.Fs, latestTag, proposedTag, changelogParentPath string) (*Changelog, error) {
	changelogPath := filepath.Join(changelogParentPath, ChangelogDirectory, proposedTag)
	files, err := afero.ReadDir(fs, changelogPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading changelog directory %s", changelogPath)
	}
	proposedVersion, err := versionutils.ParseVersion(proposedTag)
	if err != nil {
		return nil, err
	}
	changelog := Changelog{
		Version: proposedVersion,
	}
	breakingChanges := false
	for _, changelogFileInfo := range files {
		if changelogFileInfo.IsDir() {
			return nil, errors.Errorf("Unexpected directory %s in changelog directory %s", changelogFileInfo.Name(), changelogPath)
		}
		changelogFilePath := filepath.Join(changelogPath, changelogFileInfo.Name())
		if changelogFileInfo.Name() == SummaryFile {
			summary, err := afero.ReadFile(fs, changelogFilePath)
			if err != nil {
				return nil, errors.Wrapf(err, "Unable to read description file %s", changelogFilePath)
			}
			changelog.Summary = string(summary)
		} else {
			changelogFile, err := ReadChangelogFile(fs, changelogFilePath)
			if err != nil {
				return nil, err
			}
			changelog.Files = append(changelog.Files, changelogFile)
			breakingChanges = breakingChanges || hasBreakingChange(changelogFile)
		}
	}
	latestVersion, err := versionutils.ParseVersion(latestTag)
	if err != nil {
		return nil, err
	}
	expectedVersion := versionutils.IncrementVersion(latestVersion, breakingChanges)
	if *proposedVersion != *expectedVersion {
		return nil, errors.Errorf("Expected version %s to be next changelog version, found %s", VersionToString(expectedVersion), VersionToString(proposedVersion))
	}
	return &changelog, nil
}
