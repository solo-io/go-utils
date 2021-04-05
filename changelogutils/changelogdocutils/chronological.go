package changelogdocutils

import (
	"context"
	"github.com/google/go-github/v32/github"
	"github.com/solo-io/go-utils/githubutils"
	"strings"
)

type ChronologicalChangelogGenerator struct {
	client    *github.Client
	repoOwner string
	repo      string
}

func NewChronologicalChangelogGenerator(client *github.Client, repoOwner, repo string) *ChronologicalChangelogGenerator {
	return &ChronologicalChangelogGenerator{
		client:    client,
		repoOwner: repoOwner,
		repo:      repo,
	}
}

func (g *ChronologicalChangelogGenerator) Generate(ctx context.Context) (string, error) {
	releases, err := githubutils.GetAllRepoReleases(ctx, g.client, g.repoOwner, g.repo)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	for _, release := range releases {
		builder.WriteString(H3(GetGithubReleaseMarkdownLink(release, g.repoOwner, g.repo)))
		builder.WriteString(release.GetBody())
	}
	return builder.String(), nil
}
