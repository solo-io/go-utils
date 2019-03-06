package changelogutils

import (
	"context"
	"github.com/google/go-github/github"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/versionutils"
	"github.com/spf13/afero"
)

type ChangelogEntry struct {
	Type            ChangelogEntryType `json:"type"`
	Description     string             `json:"description"`
	IssueLink       string             `json:"issueLink"`
	DependencyOwner string             `json:"dependencyOwner,omitempty"`
	DependencyRepo  string             `json:"dependencyRepo,omitempty"`
	DependencyTag   string             `json:"dependencyTag,omitempty"`
}

type ChangelogFile struct {
	Entries          []*ChangelogEntry `json:"changelog,omitempty"`
	ReleaseStableApi bool              `json:"releaseStableApi,omitempty"`
}

func (c *ChangelogFile) HasBreakingChange() bool {
	for _, changelogEntry := range c.Entries {
		if changelogEntry.Type == BREAKING_CHANGE {
			return true
		}
	}
	return false
}

type Changelog struct {
	Files   []*ChangelogFile
	Summary string
	Version *versionutils.Version
	Closing string
}

const (
	ChangelogDirectory = "changelog"
	SummaryFile        = "summary.md"
	ClosingFile        = "closing.md"
)

// Should return the last released version
func GetLatestTag(ctx context.Context, owner, repo string) (string, error) {
	client, err := githubutils.GetClient(ctx)
	if err != nil {
		return "", err
	}

	return githubutils.FindLatestReleaseTagIncudingPrerelease(ctx, client, owner, repo)
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
		return nil, errors.Errorf("File %s is not a valid changelog file. Error: %v", path, err)
	}

	for _, entry := range changelog.Entries {
		if entry.Type != NON_USER_FACING || entry.Type != DEPENDENCY_BUMP {
			if entry.IssueLink == "" {
				return nil, errors.Errorf("Changelog entries must have an issue link")
			}
			if entry.Description == "" {
				return nil, errors.Errorf("Changelog entries must have a description")
			}
		}
	}

	return &changelog, nil
}

func ChangelogDirExists(fs afero.Fs, changelogParentPath string) (bool, error) {
	return afero.Exists(fs, filepath.Join(changelogParentPath, ChangelogDirectory))
}

func ComputeChangelogForTag(fs afero.Fs, tag, changelogParentPath string) (*Changelog, error) {
	version, err := versionutils.ParseVersion(tag)
	if err != nil {
		return nil, err
	}
	changelog := Changelog{
		Version: version,
	}
	changelogPath := filepath.Join(changelogParentPath, ChangelogDirectory, tag)
	files, err := afero.ReadDir(fs, changelogPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading changelog directory %s", changelogPath)
	}
	for _, changelogFileInfo := range files {
		if changelogFileInfo.IsDir() {
			return nil, errors.Errorf("Unexpected directory %s in changelog directory %s", changelogFileInfo.Name(), changelogPath)
		}
		changelogFilePath := filepath.Join(changelogPath, changelogFileInfo.Name())
		if changelogFileInfo.Name() == SummaryFile {
			summary, err := afero.ReadFile(fs, changelogFilePath)
			if err != nil {
				return nil, errors.Wrapf(err, "Unable to read summary file %s", changelogFilePath)
			}
			changelog.Summary = string(summary)
		} else if changelogFileInfo.Name() == ClosingFile {
			closing, err := afero.ReadFile(fs, changelogFilePath)
			if err != nil {
				return nil, errors.Wrapf(err, "Unable to read closing file %s", changelogFilePath)
			}
			changelog.Closing = string(closing)
		} else {
			changelogFile, err := ReadChangelogFile(fs, changelogFilePath)
			if err != nil {
				return nil, err
			}
			changelog.Files = append(changelog.Files, changelogFile)
		}
	}



	return &changelog, nil
}

func ComputeChangelogForNonRelease(fs afero.Fs, latestTag, proposedTag, changelogParentPath string) (*Changelog, error) {
	latestVersion, err := versionutils.ParseVersion(latestTag)
	if err != nil {
		return nil, err
	}
	proposedVersion, err := versionutils.ParseVersion(proposedTag)
	if err != nil {
		return nil, err
	}
	if !proposedVersion.IsGreaterThan(latestVersion) {
		return nil, errors.Errorf("Proposed version %s must be greater than latest version %s", proposedVersion, latestVersion)
	}

	changelog, err := ComputeChangelogForTag(fs, proposedTag, changelogParentPath)
	if err != nil {
		return nil, err
	}
	breakingChanges := false
	releaseStableApi := false

	for _, file := range changelog.Files {
		for _, entry := range file.Entries {
			breakingChanges = breakingChanges || entry.Type.BreakingChange()
		}
		releaseStableApi = releaseStableApi || file.ReleaseStableApi
	}

	expectedVersion := latestVersion.IncrementVersion(breakingChanges)
	if releaseStableApi {
		if !proposedVersion.Equals(&versionutils.StableApiVersion) {
			return nil, errors.Errorf("Changelog indicates this is a stable API release, which should be used only to indicate the release of v1.0.0, not %s", proposedVersion)
		}
		expectedVersion = &versionutils.StableApiVersion
	}
	if *proposedVersion != *expectedVersion {
		return nil, errors.Errorf("Expected version %s to be next changelog version, found %s", expectedVersion, proposedVersion)
	}
	return changelog, nil
}

func RefHasChangelog(ctx context.Context, client *github.Client, owner, repo, sha string) (bool, error) {
	opts := &github.RepositoryContentGetOptions{
		Ref: sha,
	}

	_, branchRepoChangelog, branchResponse, err := client.Repositories.GetContents(ctx, owner, repo, ChangelogDirectory, opts)
	if err == nil && len(branchRepoChangelog) > 0 {
		return true, nil
	} else {
		if branchResponse.StatusCode != 404 {
			return false, err
		}
	}

	return false, nil
}

func TryAddDependencyChangelogs(changelog *Changelog) error {
	ctx := context.TODO()
	client, err := githubutils.GetClient(ctx)
	if err != nil {
		return err
	}
	// Add in dependency changelogs
	var dependencyChangelogs []*Changelog
	for _, changelogFile := range changelog.Files {
		for _, changelogEntry := range changelogFile.Entries {
			if changelogEntry.Type == DEPENDENCY_BUMP {
				dependencyChangelog, err := GetDependencyChangelog(ctx, client, changelogEntry.DependencyOwner, changelogEntry.DependencyRepo, changelogEntry.DependencyTag)
				if err != nil {
					return err
				}
				dependencyChangelogs = append(dependencyChangelogs, dependencyChangelog)
			}
		}
	}
	MergeDependencyChangelogs(changelog, dependencyChangelogs...)
	return nil
}

func MergeDependencyChangelogs(mainChangelog *Changelog, dependencyChangelogs ...*Changelog) {
	for _, dependencyChangelog := range dependencyChangelogs {
		for _, dependencyChangelogFile := range dependencyChangelog.Files {
			mainChangelog.Files = append(mainChangelog.Files, dependencyChangelogFile)
		}
	}
}

func GetDependencyChangelog(ctx context.Context, client *github.Client, owner, repo, tag string) (*Changelog, error) {
	version, err := versionutils.ParseVersion(tag)
	if err != nil {
		return nil, err
	}
	opts := &github.RepositoryContentGetOptions{
		Ref: tag,
	}
	directory := filepath.Join(ChangelogDirectory, tag)
	_, directoryContent, _, err := client.Repositories.GetContents(ctx, owner, repo, directory, opts)
	if err != nil {
		return nil, err
	}
	changelog := &Changelog{
		Version: version,
	}
	for _, contentFile := range directoryContent {
		content, err := contentFile.GetContent()
		if err != nil {
			return nil, err
		}
		if contentFile.GetName() == SummaryFile {
			changelog.Summary = content
		} else if contentFile.GetName() == ClosingFile {
			changelog.Closing = content
		} else {
			var changelogFile ChangelogFile
			if err := yaml.Unmarshal([]byte(content), &changelogFile); err != nil {
				return nil, errors.Errorf("Error parsing changelog file %s. Error: %v", contentFile.GetName(), err)
			}
			changelog.Files = append(changelog.Files, &changelogFile)
		}
	}
	return changelog, nil
}
