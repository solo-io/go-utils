package internal

import (
	"context"
	"time"

	"github.com/google/go-github/github"
	"github.com/solo-io/go-utils/pkgmgmtutils/brew/formula_updater_types"
)

func NewRemoteChangePusher(gitClient *github.Client) formula_updater_types.ChangePusher {
	return &remoteChangePusher{
		gitClient: gitClient,
	}
}

type remoteChangePusher struct {
	gitClient *github.Client
}

func (r *remoteChangePusher) UpdateAndPush(
	ctx context.Context,
	version string,
	versionSha string,
	branchName string,
	commitMessage string,
	perPlatformShas *formula_updater_types.PerPlatformSha256,
	formulaOptions *formula_updater_types.FormulaOptions,
) error {
	// Get original Formula contents
	// GitHub API docs: https://developer.github.com/v3/repos/contents/#get-contents
	fileContent, _, _, err := r.gitClient.Repositories.GetContents(ctx, formulaOptions.RepoOwner, formulaOptions.RepoName, formulaOptions.Path, &github.RepositoryContentGetOptions{
		Ref: "refs/heads/master",
	})
	if err != nil {
		return err
	}

	c, err := fileContent.GetContent()
	if err != nil {
		return err
	}

	// Update formula with new version information
	byt, err := UpdateFormulaBytes([]byte(c), version, versionSha, perPlatformShas, formulaOptions)
	if err != nil {
		return err
	}

	// get master branch reference
	// GitHub API docs: https://developer.github.com/v3/git/refs/#get-a-reference
	baseRef, _, err := r.gitClient.Git.GetRef(ctx, formulaOptions.RepoOwner, formulaOptions.RepoName, "refs/heads/master")
	if err != nil {
		return err
	}

	// create new branch from master branch
	// GitHub API docs: https://developer.github.com/v3/git/refs/#create-a-reference
	_, _, err = r.gitClient.Git.CreateRef(ctx, formulaOptions.RepoOwner, formulaOptions.RepoName, &github.Reference{
		Ref: github.String("refs/heads/" + branchName),
		Object: &github.GitObject{
			SHA: baseRef.Object.SHA,
		},
	})
	if err != nil {
		return err
	}

	now := time.Now()

	// git commit and push equivalent
	// GitHub API docs: https://developer.github.com/v3/repos/contents/#update-a-file
	_, _, err = r.gitClient.Repositories.UpdateFile(ctx, formulaOptions.RepoOwner, formulaOptions.RepoName, formulaOptions.Path, &github.RepositoryContentFileOptions{
		Message: github.String(commitMessage),
		Content: byt,
		SHA:     fileContent.SHA,
		Branch:  github.String(branchName),
		Committer: &github.CommitAuthor{
			Name:  github.String(formulaOptions.PRCommitName),
			Email: github.String(formulaOptions.PRCommitEmail),
			Date:  &now,
		},
	})
	if err != nil {
		return err
	}

	return nil
}
