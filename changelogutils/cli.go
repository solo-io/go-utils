package changelogutils

import (
	"context"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/spf13/afero"
)

func GetChangelogMarkdown(owner, repo string) (string, error) {
	client, err := githubutils.GetClient(context.TODO())
	if err != nil {
		return "", err
	}
	fs := afero.NewOsFs()
	latestTag, err := githubutils.FindLatestReleaseTag(context.TODO(), client, owner, repo)
	if err != nil {
		return "", err
	}
	proposedTag, err := GetProposedTag(fs, latestTag, "")
	if err != nil {
		return "", err
	}
	changelog, err := ComputeChangelog(fs, latestTag, proposedTag, "")
	if err != nil {
		return "", err
	}
	return GenerateChangelogMarkdown(changelog), nil
}
