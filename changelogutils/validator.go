package changelogutils

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/versionutils"
	"github.com/solo-io/go-utils/vfsutils"
	"strings"
)

//go:generate mockgen -destination repo_client_mock_test.go -self_package github.com/solo-io/go-utils/changelogutils -package changelogutils_test github.com/solo-io/go-utils/githubutils RepoClient


const (
	MasterBranch = "master"
)

var (
	NoChangelogFileAddedError =
		errors.Errorf("A changelog file must be added. For more information, check out https://github.com/solo-io/go-utils/tree/master/changelogutils.")
	TooManyChangelogFilesAddedError = func(filesAdded int) error {
		return errors.Errorf("Only one changelog file can be added in a PR, found %d.", filesAdded)
	}
	UnexpectedFileInChangelogDirectoryError = func(name string) error {
		return errors.Errorf("Found unexpected file %s in changelog directory.", name)
	}
	InvalidChangelogSubdirectoryNameError = func(name string) error {
		return errors.Errorf("%s is not a valid changelog directory name, must be a semver version.", name)
	}
	ListReleasesError = func(err error) error {
		return errors.Wrapf(err, "Error listing releases")
	}
	MultipleNewVersionsFoundError = func(latest, version1, version2 string) error {
		return errors.Errorf("Only one version greater than the latest release %s valid, found %s and %s.", latest, version1, version2)
	}
	NoNewVersionsFoundError = func(latest string) error {
		return errors.Errorf("No new versions greater than the latest release %s found.", latest)
	}
	AddedChangelogInOldVersionError = func(latest string) error {
		return errors.Errorf("Can only add changelog to unreleased version (currently %s)", latest)
	}
)

type ChangelogValidator interface {
	ShouldCheckChangelog(ctx context.Context) (bool, error)
	ValidateChangelog(ctx context.Context) (*ChangelogFile, error)
}

func NewChangelogValidator(client githubutils.RepoClient, code vfsutils.MountedRepo, base string) ChangelogValidator {
	return &changelogValidator{
		client: client,
		code: code,
		base: base,
	}
}

type changelogValidator struct {
	base   string
	reader ChangelogReader
	client githubutils.RepoClient
	code   vfsutils.MountedRepo
}

func (c *changelogValidator) ShouldCheckChangelog(ctx context.Context) (bool, error) {
	masterHasChangelog, err := c.client.DirectoryExists(ctx, MasterBranch, ChangelogDirectory)
	if err != nil {
		return false, err
	} else if masterHasChangelog {
		return true, nil
	}

	branchHasChangelog, err := c.client.DirectoryExists(ctx, c.code.GetSha(), ChangelogDirectory)
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
	if !strings.HasPrefix(commitFile.GetFilename(), fmt.Sprintf("%s/%s", ChangelogDirectory, proposedTag)) {
		return nil, AddedChangelogInOldVersionError(proposedTag)
	}

	return newChangelogFile, nil
}

func (c *changelogValidator) validateProposedTag(ctx context.Context) (string, error) {
	latestTag, err := c.client.FindLatestReleaseTagIncudingPrerelease(ctx)
	if err != nil {
		return "", ListReleasesError(err)
	}
	children, err := c.code.ListFiles(ctx, ChangelogDirectory)
	if err != nil {
		return "", err
	}
	proposedVersion := ""
	for _, child := range children {
		if !child.IsDir() {
			return "", UnexpectedFileInChangelogDirectoryError(child.Name())
		}
		if !versionutils.MatchesRegex(child.Name()) {
			return "", InvalidChangelogSubdirectoryNameError(child.Name())
		}
		greaterThan, err := versionutils.IsGreaterThanTag(child.Name(), latestTag)
		if err != nil {
			return "", err
		}
		if greaterThan {
			if proposedVersion != "" {
				return "", MultipleNewVersionsFoundError(latestTag, proposedVersion, child.Name())
			}
			proposedVersion = child.Name()
		}
	}
	if proposedVersion == "" {
		return "", NoNewVersionsFoundError(latestTag)
	}
	_, err = NewChangelogReader(c.code).GetChangelogForTag(ctx, proposedVersion)
	return proposedVersion, err
}



func (c *changelogValidator) validateChangelogInPr(ctx context.Context) (*github.CommitFile, *ChangelogFile, error) {
	commitComparison, err := c.client.CompareCommits(ctx, c.base, c.code.GetSha())
	if err != nil {
		return nil, nil, err
	}
	var changelogFiles []github.CommitFile
	for _, file := range commitComparison.Files {
		if strings.HasPrefix(file.GetFilename(), fmt.Sprintf("%s/", ChangelogDirectory)) {
			if file.GetStatus() == githubutils.COMMIT_FILE_STATUS_ADDED {
				changelogFiles = append(changelogFiles, file)
			}
		}
	}
	if len(changelogFiles) == 0 {
		return nil, nil, NoChangelogFileAddedError
	} else if len(changelogFiles) > 1 {
		return nil, nil, TooManyChangelogFilesAddedError(len(changelogFiles))
	}
	parsedChangelog, err := NewChangelogReader(c.code).ReadChangelogFile(ctx, changelogFiles[0].GetFilename())
	return &changelogFiles[0], parsedChangelog, err
}
