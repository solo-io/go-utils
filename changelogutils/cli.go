package changelogutils

import (
	"context"

	"github.com/solo-io/go-utils/githubutils"
	"github.com/spf13/afero"
)

func GetChangelogMarkdownForPR(owner, repo string) (string, error) {
	client, err := githubutils.GetClient(context.Background())
	if err != nil {
		return "", err
	}
	fs := afero.NewOsFs()
	latestTag, err := githubutils.FindLatestReleaseTagIncudingPrerelease(context.Background(), client, owner, repo)
	if err != nil {
		return "", err
	}
	proposedTag, err := GetProposedTag(fs, latestTag, "")
	if err != nil {
		return "", err
	}
	changelog, err := ComputeChangelogForNonRelease(fs, latestTag, proposedTag, "")
	if err != nil {
		return "", err
	}
	return GenerateChangelogMarkdown(changelog), nil
}
