package changelogutils

import (
	"context"
	"github.com/solo-io/go-utils/githubutils"
)

func GetChangelogMarkdownForPR(owner, repo string) (string, error) {
	ctx := context.TODO()
	client, err := githubutils.GetClient(context.TODO())
	if err != nil {
		return "", err
	}
	latestTag, err := githubutils.FindLatestReleaseTagIncudingPrerelease(context.TODO(), client, owner, repo)
	if err != nil {
		return "", err
	}
	reader, err := NewChangelogReader(ctx)
	if err != nil {
		return "", err
	}
	changelog, err := reader.ReadChangelogForTag(owner, repo, "master", latestTag)
	if err != nil {
		return "", err
	}
	renderer := NewDefaultChangelogRenderer()
	return renderer.Render(changelog), nil
}

