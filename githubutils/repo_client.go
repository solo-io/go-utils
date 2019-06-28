package githubutils

import (
	"context"
	"github.com/google/go-github/github"
)

type RepoClient interface {
	FindLatestReleaseTagIncudingPrerelease(ctx context.Context) (string, error)
	CompareCommits(ctx context.Context, base, sha string) (*github.CommitsComparison, error)
	DirectoryExists(ctx context.Context, sha, directory string) (bool, error)
}

type repoClient struct {
	client *github.Client
	owner string
	repo string
}

func NewRepoClient(client *github.Client, owner, repo string) RepoClient {
	return &repoClient{
		client: client,
		owner: owner,
		repo: repo,
	}
}

func (c *repoClient) FindLatestReleaseTagIncudingPrerelease(ctx context.Context) (string, error) {
	return FindLatestReleaseTagIncudingPrerelease(ctx, c.client, c.owner, c.repo)
}

func (c *repoClient) CompareCommits(ctx context.Context, base, sha string) (*github.CommitsComparison, error) {
	commitComparison, _, err := c.client.Repositories.CompareCommits(ctx, c.owner, c.repo, base, sha)
	if err != nil {
		return nil, err
	}
	return commitComparison, err
}

func (c *repoClient) DirectoryExists(ctx context.Context, sha, directory string) (bool, error) {
	opts := &github.RepositoryContentGetOptions{
		Ref: sha,
	}
	_, repoDirectory, branchResponse, err := c.client.Repositories.GetContents(ctx, c.owner, c.repo, directory, opts)
	if err == nil && len(repoDirectory) > 0 {
		return true, nil
	} else {
		if branchResponse.StatusCode != 404 {
			return false, err
		}
	}
	return false, nil
}


