package githubutils

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"

	"github.com/google/go-github/github"
)

type PRSpec struct {
	Message string
}

type RepoClient interface {
	FindLatestReleaseTagIncudingPrerelease(ctx context.Context) (string, error)
	CompareCommits(ctx context.Context, base, sha string) (*github.CommitsComparison, error)
	DirectoryExists(ctx context.Context, sha, directory string) (bool, error)
	CreateBranch(ctx context.Context, branchName string) (*github.Reference, error)
	CreatePR(ctx context.Context, branchName string, spec PRSpec) error
	GetShaForTag(ctx context.Context, tag string) (string, error)
}

type repoClient struct {
	client *github.Client
	owner  string
	repo   string
}

func NewRepoClient(client *github.Client, owner, repo string) RepoClient {
	return &repoClient{
		client: client,
		owner:  owner,
		repo:   repo,
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

func (c *repoClient) CreateBranch(ctx context.Context, branchName string) (*github.Reference, error) {
	// get master branch reference
	// GitHub API docs: https://developer.github.com/v3/git/refs/#get-a-reference
	masterRef, _, err := c.client.Git.GetRef(ctx, c.owner, c.repo, "refs/heads/master")
	if err != nil {
		return nil, err
	}

	// create new branch from master branch
	// GitHub API docs: https://developer.github.com/v3/git/refs/#create-a-reference
	ref, _, err := c.client.Git.CreateRef(ctx, c.owner, c.repo, &github.Reference{
		Ref: github.String("refs/heads/" + branchName),
		Object: &github.GitObject{
			SHA: masterRef.Object.SHA,
		},
	})
	if err != nil {
		return nil, err
	}
	return ref, nil
}

func (c *repoClient) CreatePR(ctx context.Context, branchName string, spec PRSpec) error {
	newPR := &github.NewPullRequest{
		Title:               github.String(spec.Message),
		Head:                github.String(branchName),
		Base:                github.String("master"),
		Body:                github.String(spec.Message),
		MaintainerCanModify: github.Bool(true),
	}
	pr, _, err := c.client.PullRequests.Create(ctx, c.owner, c.repo, newPR)
	if err != nil {
		return err
	}
	contextutils.LoggerFrom(ctx).Infow("PR created",
		zap.String("url", pr.GetHTMLURL()))
	return nil
}

func (c *repoClient) GetShaForTag(ctx context.Context, tag string) (string, error) {
	ref, _, err := c.client.Git.GetRef(ctx, c.owner, c.owner, "tags/"+tag)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error loading ref for tag", zap.Error(err), zap.String("tag", tag))
		return "", err
	}
	return *ref.Object.SHA, nil
}
