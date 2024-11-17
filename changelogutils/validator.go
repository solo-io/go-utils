package changelogutils

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/stringutils"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/versionutils"
	"github.com/solo-io/go-utils/vfsutils"
)

//go:generate mockgen -destination repo_client_mock_test.go -self_package github.com/solo-io/go-utils/changelogutils -package changelogutils_test github.com/solo-io/go-utils/githubutils RepoClient

const (
	MasterBranch           = "master"
	ValidationSettingsFile = "validation.yaml"
)

var (
	knownFiles = []string{
		GetValidationSettingsPath(),
	}

	defaultSettings = ValidationSettings{}

	NoChangelogFileAddedError       = eris.Errorf("A changelog file must be added. For more information, check out https://github.com/solo-io/go-utils/tree/master/changelogutils.")
	TooManyChangelogFilesAddedError = func(filesAdded int) error {
		return eris.Errorf("Only one changelog file can be added in a PR, found %d.", filesAdded)
	}
	UnexpectedFileInChangelogDirectoryError = func(name string) error {
		return eris.Errorf("Found unexpected file %s in changelog directory.", name)
	}
	InvalidChangelogSubdirectoryNameError = func(name string) error {
		return eris.Errorf("%s is not a valid changelog directory name, must be a semver version.", name)
	}
	ListReleasesError = func(err error) error {
		return errors.Wrapf(err, "Error listing releases")
	}
	MultipleNewVersionsFoundError = func(latest, version1, version2 string) error {
		return eris.Errorf("Only one version greater than the latest release %s valid, found %s and %s.", latest, version1, version2)
	}
	NoNewVersionsFoundError = func(latest string) error {
		return eris.Errorf("No new versions greater than the latest release %s found.", latest)
	}
	AddedChangelogInOldVersionError = func(latest string) error {
		return eris.Errorf("Can only add changelog to unreleased version (currently %s)", latest)
	}
	InvalidUseOfStableApiError = func(tag string) error {
		return eris.Errorf("Changelog indicates this is a stable API release, which should be used only to indicate the release of v1.0.0, not %s", tag)
	}
	UnexpectedProposedVersionError = func(expected, actual string) error {
		return eris.Errorf("Expected version %s to be next changelog version, found %s", expected, actual)
	}
	UnableToGetSettingsError = func(err error) error {
		return errors.Wrapf(err, "Unable to read settings file")
	}
	InvalidLabelError = func(label string, allowed []string) error {
		return eris.Errorf("Changelog version has label %s, which isn't in the list of allowed labels: %v", label, allowed)
	}
	ExpectedVersionLabelError = func(actual string) error {
		return eris.Errorf("Expected version %s to to have a semver label suffix", actual)
	}
)

type ChangelogValidator interface {
	ShouldCheckChangelog(ctx context.Context) (bool, error)
	ValidateChangelog(ctx context.Context) (*ChangelogFile, error)
}

type TagComparator func(greaterThanTag, lessThanTag string) (bool, bool, error)

func NewChangelogValidatorWithLabelOrder(client githubutils.RepoClient, code vfsutils.MountedRepo, base string, labelOrder []string) ChangelogValidator {
	return &changelogValidator{
		client:     client,
		code:       code,
		base:       base,
		labelOrder: labelOrder,
	}
}

func NewChangelogValidator(client githubutils.RepoClient, code vfsutils.MountedRepo, base string) ChangelogValidator {
	return NewChangelogValidatorWithLabelOrder(client, code, base, nil)
}

type ValidationSettings struct {
	// If true, then the validator will skip checks to enforce how version numbers are incremented, allowing for more flexible
	// versioning for new features or breaking changes
	RelaxSemverValidation bool `json:"relaxSemverValidation"`

	// If true, then the validator will require a changelog version with a label.
	// This is useful to enforce version schemes like we use for envoy-gloo / envoy-gloo-ee, which always have the form:
	// $ENVOY_VERSION-$PATCH_NUM
	RequireLabel bool `json:"requireLabel"`

	// If non-empty, then the validator will reject a changelog if the version's label is not contained in this slice
	AllowedLabels []string `json:"allowedLabels"`

	// defaults to "".  When set, allows for a nested processing schema.  ex: "v1.10" would mean only files in "changelog/v1.10" would be processed
	ActiveSubdirectory string `json:"activeSubdirectory"`
}

type changelogValidator struct {
	base   string
	reader ChangelogReader
	client githubutils.RepoClient
	code   vfsutils.MountedRepo
	// list of arbitrary labels whose order is used to tie-break tag comparisons between
	// versions with different labels. Labels ordered earlier are greater.
	labelOrder []string
}

func (c *changelogValidator) ShouldCheckChangelog(ctx context.Context) (bool, error) {
	dir := c.GetChangelogDirectory(ctx)
	masterHasChangelog, err := c.client.DirectoryExists(ctx, MasterBranch, dir)
	if err != nil {
		return false, err
	} else if masterHasChangelog {
		return true, nil
	}

	branchHasChangelog, err := c.client.DirectoryExists(ctx, c.code.GetSha(), dir)
	if err != nil {
		return false, err
	}
	return branchHasChangelog, nil
}

func (c *changelogValidator) ValidateChangelog(ctx context.Context) (*ChangelogFile, error) {
	check, err := c.ShouldCheckChangelog(ctx)
	if err != nil {
		return nil, err
	} else if !check {
		return nil, nil
	}

	commitFile, newChangelogFile, err := c.validateChangelogInPr(ctx)
	if err != nil {
		return nil, err
	}

	proposedTag, err := c.validateProposedTag(ctx)
	if err != nil {
		return nil, err
	}

	// validate commit file for tag
	if !strings.HasPrefix(commitFile.GetFilename(), fmt.Sprintf("%s/%s", c.GetChangelogDirectory(ctx), proposedTag)) {
		return nil, AddedChangelogInOldVersionError(proposedTag)
	}

	return newChangelogFile, nil
}

func (c *changelogValidator) validateProposedTag(ctx context.Context) (string, error) {
	latestTag, err := c.client.FindLatestTagIncludingPrereleaseBeforeSha(ctx, c.base)
	if err != nil {
		return "", ListReleasesError(err)
	}

	dir := c.GetChangelogDirectory(ctx)
	children, err := c.code.ListFiles(ctx, dir)
	if err != nil {
		return "", err
	}
	proposedVersion := ""
	for _, child := range children {
		if !child.IsDir() {
			if !IsKnownChangelogFile(filepath.Join(dir, child.Name())) {
				return "", UnexpectedFileInChangelogDirectoryError(child.Name())
			} else {
				continue
			}
		}
		if !versionutils.MatchesRegex(child.Name()) {
			return "", InvalidChangelogSubdirectoryNameError(child.Name())
		}
		var greaterThan, determinable bool
		if len(c.labelOrder) > 0 {
			greaterThan, determinable, err = versionutils.IsGreaterThanTagWithLabelOrder(child.Name(), latestTag, c.labelOrder)

		} else {
			greaterThan, determinable, err = versionutils.IsGreaterThanTag(child.Name(), latestTag)

		}
		if err != nil {
			return "", err
		}
		if greaterThan || !determinable {
			if proposedVersion != "" {
				return "", MultipleNewVersionsFoundError(latestTag, proposedVersion, child.Name())
			}
			proposedVersion = child.Name()
		}
	}
	if proposedVersion == "" {
		return "", NoNewVersionsFoundError(latestTag)
	}
	changelog, err := NewChangelogReader(c.code).GetChangelogForTag(ctx, proposedVersion)
	if err != nil {
		return proposedVersion, err
	}
	err = c.validateVersionBump(ctx, latestTag, changelog)
	return proposedVersion, err
}

func (c *changelogValidator) validateVersionBump(ctx context.Context, latestTag string, changelog *Changelog) error {
	latestVersion, err := versionutils.ParseVersion(latestTag)
	if err != nil {
		return err
	}

	breakingChanges := false
	newFeature := false
	releaseStableApi := false

	// get settings now to ensure this function returns an error on invalid settings
	settings, err := c.getValidationSettings(ctx)
	if err != nil {
		// validation settings should be defined in a "validation.yaml" file, or fall back to default.
		// if an error is returned, that means there was a settings file, but it was malformed, and we
		// propagate such an error to ensure the branch stays clean
		return err
	}

	if settings.RequireLabel && len(changelog.Version.Label) == 0 {
		return ExpectedVersionLabelError(changelog.Version.String())
	}

	// If the settings contain specific allowed labels, ensure the label used here, if any, is in the list
	if changelog.Version.Label != "" && len(settings.AllowedLabels) > 0 {
		if !stringutils.ContainsString(changelog.Version.Label, settings.AllowedLabels) {
			return InvalidLabelError(changelog.Version.Label, settings.AllowedLabels)
		}
	}

	for _, file := range changelog.Files {
		for _, entry := range file.Entries {
			breakingChanges = breakingChanges || entry.Type.BreakingChange()
			newFeature = newFeature || entry.Type.NewFeature()
		}
		releaseStableApi = releaseStableApi || file.GetReleaseStableApi()
	}

	// this flag can be used in the changelog to signal a stable release, which could be 1.0.0 or 1.5.0 or X.Y.0
	if releaseStableApi {
		// if the changelog is less than 1.0, then this isn't a stable API
		if !changelog.Version.MustIsGreaterThanOrEqualTo(versionutils.StableApiVersion()) {
			return InvalidUseOfStableApiError(changelog.Version.String())
		}

		// if this is supposed to be a stable release, then the patch and release candidate
		// versions should be 0. This enables release histories like:
		// 0.10 -> 1.0.0-rc1 -> 1.0.0-rc2 -> 1.0.0 -> 1.1.0-rc1 -> 1.1.0 -> ...
		if changelog.Version.Patch != 0 || changelog.Version.LabelVersion != 0 {
			return InvalidUseOfStableApiError(changelog.Version.String())
		}
		return nil
	}

	expectedVersion := latestVersion.IncrementVersion(breakingChanges, newFeature)
	// if this isn't the first labeled version, and we aren't switching label versions (e.g. 1.0.0-beta1 -> 1.0.0-rc1)
	// then the version should match the expected version exactly
	if changelog.Version.LabelVersion > 1 && changelog.Version.Label == expectedVersion.Label && !changelog.Version.Equals(expectedVersion) {
		return UnexpectedProposedVersionError(expectedVersion.String(), changelog.Version.String())
	}

	if changelog.Version.Label == "" && expectedVersion.Label == "" && !settings.RelaxSemverValidation {
		// since this isn't a labeled release or a stable release, the version should be incremented
		// based on semver rules.
		if changelog.Version.LabelVersion == 0 && !changelog.Version.Equals(expectedVersion) {
			return UnexpectedProposedVersionError(expectedVersion.String(), changelog.Version.String())
		}
	}

	return nil
}

func (c *changelogValidator) validateChangelogInPr(ctx context.Context) (*github.CommitFile, *ChangelogFile, error) {
	changelogFiles, err := GetChangelogFilesAdded(ctx, c.client, c.base, c.code.GetSha())
	if err != nil {
		return nil, nil, err
	}
	if len(changelogFiles) == 0 {
		return nil, nil, NoChangelogFileAddedError
	} else if len(changelogFiles) > 1 {
		return nil, nil, TooManyChangelogFilesAddedError(len(changelogFiles))
	}
	parsedChangelog, err := NewChangelogReader(c.code).ReadChangelogFile(ctx, changelogFiles[0].GetFilename())
	return &changelogFiles[0], parsedChangelog, err
}

func GetChangelogFilesAdded(ctx context.Context, client githubutils.RepoClient, base, sha string) ([]github.CommitFile, error) {
	commitComparison, err := client.CompareCommits(ctx, base, sha)
	if err != nil {
		return nil, err
	}
	var changelogFiles []github.CommitFile
	for _, file := range commitComparison.Files {
		// leaving ChangelogDirectory hardcoded to "changelog" is a non-issue here, since we are lazy prefix matching against "changelog/*"
		if strings.HasPrefix(file.GetFilename(), fmt.Sprintf("%s/", ChangelogDirectory)) {
			if !IsKnownChangelogFile(file.GetFilename()) && file.GetStatus() == githubutils.COMMIT_FILE_STATUS_ADDED {
				changelogFiles = append(changelogFiles, *file)
			}
		}
	}
	return changelogFiles, nil
}

func GetValidationSettingsPath() string {
	// leaving ChangelogDirectory hardcoded to "changelog" is a non-issue here, since even _if_  ActiveSubdirectory is set, we still only
	// want to consider 1 (top-level) settings file
	return fmt.Sprintf("%s/%s", ChangelogDirectory, ValidationSettingsFile)
}

func IsKnownChangelogFile(path string) bool {
	return stringutils.ContainsString(path, knownFiles)
}

func (c *changelogValidator) getValidationSettings(ctx context.Context) (*ValidationSettings, error) {
	return GetValidationSettings(ctx, c.code, c.client)
}

func (c *changelogValidator) GetChangelogDirectory(ctx context.Context) string {
	// as a potential optimization, we could make ValidationSettings a singleton to prevent continually reading from a remote GH
	// during a single validation operation.  I've elected to _not_ do so here, since I'm not totally sure if `changelogValidator`'s are
	// used in a "1 and done" capacity.  Calling this out "in case"
	settings, _ := c.getValidationSettings(ctx) // suppressing error, because we _should_ always know the changelog dir

	if settings != nil && settings.ActiveSubdirectory != "" {
		return "changelog/" + settings.ActiveSubdirectory
	}
	return "changelog"
}

func GetValidationSettings(ctx context.Context, code vfsutils.MountedRepo, client githubutils.RepoClient) (*ValidationSettings, error) {
	exists, err := client.FileExists(ctx, code.GetSha(), GetValidationSettingsPath())
	if err != nil {
		return nil, UnableToGetSettingsError(err)
	}
	if !exists {
		return &defaultSettings, nil
	}

	var settings ValidationSettings
	bytes, err := code.GetFileContents(ctx, GetValidationSettingsPath())
	if err != nil {
		return nil, UnableToGetSettingsError(err)
	}

	if err := yaml.Unmarshal(bytes, &settings); err != nil {
		return nil, UnableToGetSettingsError(err)
	}
	return &settings, nil
}
