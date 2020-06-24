package internal

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/solo-io/go-utils/pkgmgmtutils/formula_updater_types"
)

// uses real dependencies- i.e. actually communicates with GitHub
func NewGitClient(client *github.Client) formula_updater_types.GitClient {
	return &gitClient{
		client: client,
	}
}

type gitClient struct {
	client *github.Client
}

func (g *gitClient) GetRefSha(ctx context.Context, owner string, repo string, ref string) (string, error) {
	foundRef, _, err := g.client.Git.GetRef(ctx, owner, repo, ref)
	if err != nil {
		return "", err
	}

	return *foundRef.Object.SHA, nil
}

func (g *gitClient) GetReleaseAssetsByTag(ctx context.Context, owner, repo, version string) ([]formula_updater_types.ReleaseAsset, error) {
	rr, _, err := g.client.Repositories.GetReleaseByTag(ctx, owner, repo, version)
	if err != nil {
		return nil, err
	}

	var releaseAssets []formula_updater_types.ReleaseAsset
	for _, asset := range rr.Assets {
		releaseAssets = append(releaseAssets, formula_updater_types.ReleaseAsset{
			Name:               asset.GetName(),
			BrowserDownloadUrl: asset.GetBrowserDownloadURL(),
		})
	}
	return releaseAssets, nil
}

func (g *gitClient) CreatePullRequest(
	ctx context.Context,
	formulaOptions *formula_updater_types.FormulaOptions,
	commitMessage string,
	branchName string,
) error {
	prRepoOwner := formulaOptions.PRRepoOwner
	prRepoName := formulaOptions.PRRepoName

	prHead := branchName
	if prRepoOwner == "" {
		prRepoOwner = formulaOptions.RepoOwner
		prRepoName = formulaOptions.RepoName
	} else if prRepoOwner != formulaOptions.RepoOwner {
		// For cross-repo PR, prHead should be in format of "<change repo owner>:branch"
		prHead = formulaOptions.RepoOwner + ":" + branchName
	}

	base := formulaOptions.PRBranch
	if base == "" {
		base = "master"
	}

	// Create GitHub Pull Request
	// GitHub API docs: https://developer.github.com/v3/pulls/#create-a-pull-request
	_, _, err := g.client.PullRequests.Create(ctx, prRepoOwner, prRepoName, &github.NewPullRequest{
		Title:               github.String(commitMessage),
		Head:                github.String(prHead),
		Base:                github.String(base),
		Body:                github.String(formulaOptions.PRDescription),
		MaintainerCanModify: github.Bool(true),
	})
	return err
}
