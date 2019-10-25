package githubutils

import (
	"context"

	"github.com/solo-io/go-utils/versionutils"

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
	FileExists(ctx context.Context, sha, path string) (bool, error)
	CreateBranch(ctx context.Context, branchName string) (*github.Reference, error)
	CreatePR(ctx context.Context, branchName string, spec PRSpec) error
	GetShaForTag(ctx context.Context, tag string) (string, error)
	GetPR(ctx context.Context, num int) (*github.PullRequest, error)
	UpdateRelease(ctx context.Context, release *github.RepositoryRelease) (*github.RepositoryRelease, error)
	GetCommit(ctx context.Context, sha string) (*github.RepositoryCommit, error)
	FindStatus(ctx context.Context, statusLabel, sha string) (*github.RepoStatus, error)
	CreateStatus(ctx context.Context, sha string, status *github.RepoStatus) (*github.RepoStatus, error)
	CreateComment(ctx context.Context, pr int, comment *github.IssueComment) (*github.IssueComment, error)
	DeleteComment(ctx context.Context, commentId int64) error
	FindLatestTagIncludingPrereleaseBeforeSha(ctx context.Context, sha string) (string, error)
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

func (c *repoClient) FindLatestTagIncludingPrereleaseBeforeSha(ctx context.Context, sha string) (string, error) {
	releases, _, err := c.client.Repositories.ListReleases(ctx, c.owner, c.repo, &github.ListOptions{})
	if err != nil {
		return "", err
	}
	latestBeforeSha := versionutils.NewVersion(0, 0, 1)
	// TODO: this will eventually need to page through all releases, not just the first page
	for _, release := range releases {
		if release.GetDraft() {
			continue
		}

		comparison, _, err := c.client.Repositories.CompareCommits(ctx, c.owner, c.repo, release.GetTagName(), sha)
		if err != nil {
			return "", err
		}
		if comparison.GetStatus() == "ahead" || comparison.GetStatus() == "identical" {
			releaseVersion, err := versionutils.ParseVersion(release.GetTagName())
			if err != nil {
				return "", err
			}
			if releaseVersion.IsGreaterThan(latestBeforeSha) {
				latestBeforeSha = releaseVersion
				break
			}
		}
	}
	return latestBeforeSha.String(), nil
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
		if branchResponse != nil && branchResponse.StatusCode != 404 {
			contextutils.LoggerFrom(ctx).Errorw("Unable to determine whether ref has directory",
				zap.Error(err),
				zap.String("sha", sha),
				zap.String("directory", directory))
			return false, err
		}
	}
	return false, nil
}

func (c *repoClient) FileExists(ctx context.Context, sha, path string) (bool, error) {
	opts := &github.RepositoryContentGetOptions{
		Ref: sha,
	}
	_, _, branchResponse, err := c.client.Repositories.GetContents(ctx, c.owner, c.repo, path, opts)
	if err == nil {
		return true, nil
	} else {
		if branchResponse != nil && branchResponse.StatusCode != 404 {
			contextutils.LoggerFrom(ctx).Errorw("Unable to determine whether ref has file",
				zap.Error(err),
				zap.String("sha", sha),
				zap.String("path", path))
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
	ref, _, err := c.client.Git.GetRef(ctx, c.owner, c.repo, "tags/"+tag)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error loading ref for tag", zap.Error(err), zap.String("tag", tag))
		return "", err
	}
	return *ref.Object.SHA, nil
}

func (c *repoClient) GetPR(ctx context.Context, num int) (*github.PullRequest, error) {
	pr, _, err := c.client.PullRequests.Get(ctx, c.owner, c.repo, num)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("can't get PR object",
			zap.Error(err),
			zap.String("owner", c.owner),
			zap.String("repo", c.repo),
			zap.Int("prNumber", num))
		return nil, err
	}
	return pr, nil
}

func (c *repoClient) UpdateRelease(ctx context.Context, release *github.RepositoryRelease) (*github.RepositoryRelease, error) {
	updatedRelease, _, err := c.client.Repositories.EditRelease(ctx, c.owner, c.repo, release.GetID(), release)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Unable to update release", zap.Error(err))
		return nil, err
	}
	return updatedRelease, nil
}

func (c *repoClient) GetCommit(ctx context.Context, sha string) (*github.RepositoryCommit, error) {
	commit, _, err := c.client.Repositories.GetCommit(ctx, c.owner, c.repo, sha)
	return commit, err
}

func (c *repoClient) FindStatus(ctx context.Context, statusLabel, sha string) (*github.RepoStatus, error) {
	return FindStatus(ctx, c.client, statusLabel, c.owner, c.repo, sha)
}

func (c *repoClient) CreateStatus(ctx context.Context, sha string, status *github.RepoStatus) (*github.RepoStatus, error) {
	// truncate if necessary
	if len(status.GetDescription()) > 140 {
		status.Description = github.String(status.GetDescription()[:140])
	}
	st, _, err := c.client.Repositories.CreateStatus(ctx, c.owner, c.repo, sha, status)
	return st, err
}

func (c *repoClient) CreateComment(ctx context.Context, pr int, comment *github.IssueComment) (*github.IssueComment, error) {
	created, _, err := c.client.Issues.CreateComment(ctx, c.owner, c.repo, pr, comment)
	return created, err
}

func (c *repoClient) DeleteComment(ctx context.Context, commentId int64) error {
	_, err := c.client.Issues.DeleteComment(ctx, c.owner, c.repo, commentId)
	return err
}
