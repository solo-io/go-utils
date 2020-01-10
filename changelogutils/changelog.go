package changelogutils

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/google/go-github/github"
	"github.com/rotisserie/eris"

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
	ResolvesIssue   *bool              `json:"resolvesIssue,omitempty"`
}

func (c *ChangelogEntry) GetResolvesIssue() bool {
	if c.ResolvesIssue == nil {
		return true
	}
	return *c.ResolvesIssue
}

type ChangelogFile struct {
	Entries          []*ChangelogEntry `json:"changelog,omitempty"`
	ReleaseStableApi *bool             `json:"releaseStableApi,omitempty"`
}

func (c *ChangelogFile) GetReleaseStableApi() bool {
	if c.ReleaseStableApi == nil {
		return false
	}
	return *c.ReleaseStableApi
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

type errorNoVersionFound struct {
	version string
}

func newErrorNoVersionFound(version string) *errorNoVersionFound {
	return &errorNoVersionFound{version: version}
}

func (e *errorNoVersionFound) Error() string {
	return fmt.Sprintf("No version greater than %s found", e.version)
}

func IsNoVersionFoundError(err error) bool {
	_, ok := err.(*errorNoVersionFound)
	return ok
}

type errorMultipleVersionsFound struct {
	changelogFile, proposedVersion, latestTag string
}

func newErrorMultipleVersionsFound(changelogFile string, proposedVersion string, latestTag string) *errorMultipleVersionsFound {
	return &errorMultipleVersionsFound{changelogFile: changelogFile, proposedVersion: proposedVersion, latestTag: latestTag}
}

func (e *errorMultipleVersionsFound) Error() string {
	return fmt.Sprintf("Versions %s and %s are both greater than latest tag %s", e.changelogFile, e.proposedVersion, e.latestTag)
}

func IsMultipleVersionsFoundError(err error) bool {
	_, ok := err.(*errorMultipleVersionsFound)
	return ok
}

type errorInvalidDirectoryName struct {
	dir string
}

func newErrorInvalidDirectoryName(dir string) *errorInvalidDirectoryName {
	return &errorInvalidDirectoryName{dir: dir}
}

func IsInvalidDirectoryNameError(err error) bool {
	_, ok := err.(*errorInvalidDirectoryName)
	return ok
}

func (e *errorInvalidDirectoryName) Error() string {
	return fmt.Sprintf("Directory name %s is not valid, must be of the form 'vX.Y.Z'", e.dir)
}

// Should return the last released version
// Deprecated: use githubutils.RepoClient.FindLatestReleaseIncludingPrerelease instead
func GetLatestTag(ctx context.Context, owner, repo string) (string, error) {
	client, err := githubutils.GetClient(ctx)
	if err != nil {
		return "", err
	}

	return githubutils.FindLatestReleaseTagIncudingPrerelease(ctx, client, owner, repo)
}

// Deprecated: use ChangelogValidator instead
func GetProposedTagForRepo(ctx context.Context, client *github.Client, owner, repo string) (string, error) {
	latestTag, err := githubutils.FindLatestReleaseTagIncudingPrerelease(ctx, client, owner, repo)
	if err != nil {
		return "", err
	}
	_, changelogContents, _, err := client.Repositories.GetContents(ctx, owner, repo, ChangelogDirectory, nil)
	if err != nil {
		return "", err
	}
	var proposedVersion string
	for _, changelogFile := range changelogContents {
		if changelogFile.GetType() != githubutils.CONTENT_TYPE_DIRECTORY {
			continue
		}

		if !versionutils.MatchesRegex(changelogFile.GetName()) {
			return "", newErrorInvalidDirectoryName(changelogFile.GetName())
		}
		greaterThan, determinable, err := versionutils.IsGreaterThanTag(changelogFile.GetName(), latestTag)
		if err != nil {
			return "", err
		}
		if greaterThan || !determinable {
			if proposedVersion != "" {
				return "", newErrorMultipleVersionsFound(changelogFile.GetName(), proposedVersion, latestTag)
			}
			proposedVersion = changelogFile.GetName()
		}
	}
	if proposedVersion == "" {
		return "", newErrorNoVersionFound(latestTag)
	}
	return proposedVersion, nil
}

// Should return the next version to release, based on the names of the subdirectories in the changelog
// Will return an error if there is no version, or multiple versions, larger than the latest tag,
// according to semver
// Deprecated: use ChangelogValidator instead
func GetProposedTag(fs afero.Fs, latestTag, changelogParentPath string) (string, error) {
	// handle special case where this is the first release
	if latestTag == versionutils.SemverNilVersionValue {
		return versionutils.SemverMinimumVersion, nil
	}
	changelogPath := filepath.Join(changelogParentPath, ChangelogDirectory)
	subDirs, err := afero.ReadDir(fs, changelogPath)
	if err != nil {
		return "", errors.Wrapf(err, "Error reading changelog directory")
	}
	proposedVersion := ""
	for _, subDir := range subDirs {
		if !subDir.IsDir() {
			return "", eris.Errorf("Unexpected entry %s in changelog directory", subDir.Name())
		}
		if !versionutils.MatchesRegex(subDir.Name()) {
			return "", newErrorInvalidDirectoryName(subDir.Name())
		}
		greaterThan, determinable, err := versionutils.IsGreaterThanTag(subDir.Name(), latestTag)
		if err != nil {
			return "", err
		}
		if greaterThan || !determinable {
			if proposedVersion != "" {
				return "", newErrorMultipleVersionsFound(subDir.Name(), proposedVersion, latestTag)
			}
			proposedVersion = subDir.Name()
		}
	}
	if proposedVersion == "" {
		return "", newErrorNoVersionFound(latestTag)
	}
	return proposedVersion, nil
}

// Deprecated: use changelogutils.ChangelogReader instead
func ReadChangelogFile(fs afero.Fs, path string) (*ChangelogFile, error) {
	var changelog ChangelogFile
	bytes, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed reading changelog file: %s", path)
	}

	if err := yaml.Unmarshal(bytes, &changelog); err != nil {
		return nil, eris.Errorf("File %s is not a valid changelog file. Error: %v",
			filepath.Join(filepath.Base(filepath.Dir(path)), filepath.Base(path)), err)
	}

	for _, entry := range changelog.Entries {
		if entry.Type != NON_USER_FACING && entry.Type != DEPENDENCY_BUMP {
			if entry.IssueLink == "" {
				return nil, eris.Errorf("Changelog entries must have an issue link")
			}
			if entry.Description == "" {
				return nil, eris.Errorf("Changelog entries must have a description")
			}
		}
		if entry.Type == DEPENDENCY_BUMP {
			if entry.DependencyOwner == "" {
				return nil, eris.Errorf("Dependency bumps must have an owner")
			}
			if entry.DependencyRepo == "" {
				return nil, eris.Errorf("Dependency bumps must have a repo")
			}
			if entry.DependencyTag == "" {
				return nil, eris.Errorf("Dependency bumps must have a tag")
			}
		}
	}

	return &changelog, nil
}

// Deprecated
func ChangelogDirExists(fs afero.Fs, changelogParentPath string) (bool, error) {
	return afero.Exists(fs, filepath.Join(changelogParentPath, ChangelogDirectory))
}

// Deprecated: use changelogutils.ChangelogReader instead
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
			return nil, eris.Errorf("Unexpected directory %s in changelog directory %s", changelogFileInfo.Name(), changelogPath)
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

// Deprecated: use changelogutils.ChangelogReader instead
func ComputeChangelogForNonRelease(fs afero.Fs, latestTag, proposedTag, changelogParentPath string) (*Changelog, error) {
	latestVersion, err := versionutils.ParseVersion(latestTag)
	if err != nil {
		return nil, err
	}
	proposedVersion, err := versionutils.ParseVersion(proposedTag)
	if err != nil {
		return nil, err
	}
	isGreater, determinable, err := proposedVersion.IsGreaterThanPtr(latestVersion)
	if err != nil {
		return nil, err
	}
	if !isGreater && determinable {
		return nil, eris.Errorf("Proposed version %s must be greater than latest version %s", proposedVersion, latestVersion)
	}

	changelog, err := ComputeChangelogForTag(fs, proposedTag, changelogParentPath)
	if err != nil {
		return nil, err
	}
	breakingChanges := false
	newFeature := false
	releaseStableApi := false

	for _, file := range changelog.Files {
		for _, entry := range file.Entries {
			breakingChanges = breakingChanges || entry.Type.BreakingChange()
			newFeature = newFeature || entry.Type.NewFeature()
		}
		releaseStableApi = releaseStableApi || file.GetReleaseStableApi()
	}

	expectedVersion := latestVersion.IncrementVersion(breakingChanges, newFeature)
	if releaseStableApi {
		stableApiVer := versionutils.StableApiVersion()
		if !proposedVersion.Equals(&stableApiVer) {
			return nil, eris.Errorf("Changelog indicates this is a stable API release, which should be used only to indicate the release of v1.0.0, not %s", proposedVersion)
		}
		expectedVersion = &stableApiVer
	}
	if proposedVersion.LabelVersion == 0 && *proposedVersion != *expectedVersion {
		return nil, eris.Errorf("Expected version %s to be next changelog version, found %s", expectedVersion, proposedVersion)
	}
	return changelog, nil
}

// Deprecated: use githubutils.RepoClient.DirectoryExists
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

// Sort interface implementation for ChangelogList

type ChangelogList []*Changelog

var _ sort.Interface = ChangelogList{}

func (l ChangelogList) Len() int {
	return len(l)
}

// it is a bug to pass a changelog list containing a nil version to this function
func (l ChangelogList) Less(i, j int) bool {
	return !l[i].Version.MustIsGreaterThanOrEqualTo(*l[j].Version)
}

func (l ChangelogList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
