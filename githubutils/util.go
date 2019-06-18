package githubutils

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"
)

const (
	BranchMaster = "master"

	owner = "solo-io"
)

func CreateFullRef(ref string) string {
	return fmt.Sprintf("refs/heads/%s", ref)
}

func CreateBranch(client *github.Client, ctx context.Context, repo, branchName string) error {
	masterRef, _, err := client.Git.GetRef(ctx, owner, repo, CreateFullRef(BranchMaster))
	if err != nil {
		return err
	}
	sha1 := masterRef.Object.SHA

	newRefRequest := &github.Reference{
		Ref: github.String(branchName),
		Object: &github.GitObject{
			SHA: sha1,
		},
	}
	_, _, err = client.Git.CreateRef(ctx, owner, repo, newRefRequest)
	if err != nil {
		return err
	}
	return nil
}
