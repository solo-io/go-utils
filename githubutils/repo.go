package githubutils

import (
	"context"
	"github.com/solo-io/go-utils/errors"
	"os"

	"github.com/google/go-github/github"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

const (
	GITHUB_TOKEN = "GITHUB_TOKEN"
)

func getGithubToken() (string, error) {
	token, found := os.LookupEnv(GITHUB_TOKEN)
	if !found {
		return "", errors.Errorf("Could not find %s in environment.", GITHUB_TOKEN)
	}
	return token, nil
}

func GetClient(ctx context.Context) (*github.Client, error) {
	token, err := getGithubToken()
	if err != nil {
		return nil, err
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return client, nil
}

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
