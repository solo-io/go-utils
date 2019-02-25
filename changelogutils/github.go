package changelogutils

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/versionutils"
)

func GetProposedTagFromGit(ctx context.Context, client *github.Client, owner, repo, ref string) (string, error) {
	latestTag, err := GetLatestTag(ctx, owner, repo)
	if err != nil {
		return "", err
	}
	files, err := githubutils.GetFilesFromGit(ctx, client, owner, repo, ref, ChangelogDirectory)
	if err != nil {
		return "", nil
	}

	proposedVersion := ""
	for _, file := range files {
		if file.GetType() == githubutils.CONTENT_TYPE_DIRECTORY {
			version := file.GetName()
			if !versionutils.MatchesRegex(version) {
				return "", errors.Errorf("Directory name %s is not valid, must be of the form 'vX.Y.Z'", version)
			}

			proposedVersion, err = validVersion(version, proposedVersion, latestTag)
		}
	}
	if proposedVersion == "" {
		return "", errors.Errorf("No version greater than %s found", latestTag)
	}
	return proposedVersion, err
}

func ValidateProposedChangelogTag(ctx context.Context, client *github.Client, owner, repo, ref string) (bool, error) {
	proposedTag, err := GetProposedTagFromGit(ctx, client, owner, repo, ref)
	if err != nil {
		return false, nil
	}

	files, err := githubutils.GetFilesForChangelogVersion(ctx, client, owner, repo, ref, proposedTag)
	if err != nil {
		return false, nil
	}

	rawChangelogFiles := make([]*RawChangelogFile, 0)
	for _, file := range files {
		byt, err := githubutils.GetRawGitFile(ctx, client, file, owner, repo, ref)
		if err != nil {
			return false, err
		}
		rawChangelogFiles = append(rawChangelogFiles, &RawChangelogFile{
			Bytes: byt,
			Filename: file.GetName(),
		})
	}

	changelog, err := BytesToChangelog(rawChangelogFiles)
	if err != nil {
		return false, err
	}
	fmt.Println(changelog)
	return false, nil
}
