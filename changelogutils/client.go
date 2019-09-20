package changelogutils

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/githubutils"
	"go.uber.org/zap"
)

type ChangelogClient interface {
	GetChangelogForTag(ctx context.Context, sha, tag string) (*Changelog, error)
}

type changelogClient struct {
	client githubutils.RepoClient
}

func (c *changelogClient) GetChangelogForTag(ctx context.Context, sha, tag string) (*Changelog, error) {
	tagDir := ChangelogDirectory + "/" + tag
	exists, err := c.client.DirectoryExists(ctx, sha, tagDir)
	if err != nil || !exists {
		return nil, err
	}
	code := c.client.GetCode(ctx, sha)
	reader := NewChangelogReader(code)
	changelog, err := reader.GetChangelogForTag(ctx, tag)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error rendering changelog", zap.Error(err))
		return nil, err
	}
	return changelog, nil
}
