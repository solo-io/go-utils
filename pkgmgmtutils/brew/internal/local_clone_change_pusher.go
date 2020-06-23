package internal

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/pkgmgmtutils/brew/formula_updater_types"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

func NewLocalCloneChangePusher() formula_updater_types.ChangePusher {
	return &localCloneChangePusher{}
}

type localCloneChangePusher struct {
}

func (l *localCloneChangePusher) UpdateAndPush(
	_ context.Context,
	version string,
	versionSha string,
	branchName string,
	commitMessage string,
	perPlatformShas *formula_updater_types.PerPlatformSha256,
	formulaOptions *formula_updater_types.FormulaOptions,
) error {
	// create temp dir for local git clone
	dirTemp, err := ioutil.TempDir("", formulaOptions.RepoName)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dirTemp) // Cleanup local clone when done

	// git Clone Github repo
	repo, err := git.PlainClone(dirTemp, false, &git.CloneOptions{
		URL: "https://github.com/" + formulaOptions.RepoOwner + "/" + formulaOptions.RepoName,
	})
	if err != nil {
		return err
	}

	// Get git repo contents locally
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	// git Pull remote PR repo
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "upstream",
		URLs: []string{"https://github.com/" + formulaOptions.PRRepoOwner + "/" + formulaOptions.PRRepoName + ".git"},
	})
	if err != nil {
		return err
	}

	base := formulaOptions.PRBranch
	if base == "" {
		base = "master"
	}

	err = w.Pull(&git.PullOptions{
		RemoteName:    "upstream",
		ReferenceName: plumbing.NewBranchReferenceName(base),
	})
	if err != nil {
		return err
	}

	// git checkout -b <branchName>
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: true,
	})
	if err != nil {
		return err
	}

	formulaPath := filepath.Join(dirTemp, formulaOptions.Path)

	byt, err := ioutil.ReadFile(formulaPath)
	if err != nil {
		return err
	}

	// Update formula with new version information
	byt, err = UpdateFormulaBytes(byt, version, versionSha, perPlatformShas, formulaOptions)
	if err != nil {
		return err
	}

	// Write Updated file to git clone directory
	err = ioutil.WriteFile(formulaPath, byt, 0644)
	if err != nil {
		return err
	}

	// git commit --all
	_, err = w.Commit(commitMessage, &git.CommitOptions{
		All: true,
		Author: &object.Signature{
			Name:  formulaOptions.PRCommitName,
			Email: formulaOptions.PRCommitEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	token, err := githubutils.GetGithubToken()
	if err != nil {
		return err
	}

	// git push origin
	err = repo.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth: &http.BasicAuth{
			Username: "GitHub Token",
			Password: token,
		},
	})
	if err != nil {
		return err
	}

	return nil
}
