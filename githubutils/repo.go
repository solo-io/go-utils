package githubutils

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/rs/zerolog"
)

func FindStatus(ctx context.Context, client *github.Client, statusLabel, owner, repo, sha string) (*github.RepoStatus, error) {
	logger := zerolog.Ctx(ctx)
	statues, _, err := client.Repositories.ListStatuses(ctx, owner, repo, sha, nil)
	if err != nil {
		logger.Error().Err(err).Msg("can't list statuses")
		return nil, err
	}

	var currentStatus *github.RepoStatus
	for _, st := range statues {
		if st.Context == nil {
			continue
		}
		if *st.Context == statusLabel {
			currentStatus = st
			break
		}
	}

	return currentStatus, nil
}

func FindLatestReleaseTag(ctx context.Context, client *github.Client, owner, repo string) (string, error) {
	release, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return "", err
	}
	return *release.TagName, nil
}
