package pkgmgmtutils

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/google/go-github/github"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/versionutils"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

const PRBaseBranchDefault = "master"

type FormulaOptions struct {
	Name            string // Descriptive name to be used for logging and general identification
	FormulaName     string // proper formula name without file extension
	Path            string // repo relative path with file extension
	RepoOwner       string // repo owner for Formula change
	RepoName        string // repo name for Formula change
	PRRepoOwner     string // optional, empty means use RepoOwner
	PRRepoName      string // optional, empty means use RepoName
	PRBranch        string // optional, default to master
	PRDescription   string
	PRCommitName    string
	PRCommitEmail   string
	VersionRegex    string
	VersionShaRegex string
	DarwinShaRegex  string
	LinuxShaRegex   string
	WindowsShaRegex string

	dryRun bool
}

type FormulaStatus struct {
	Name    string
	Updated bool
	Err     error
}

func UpdateFormulas(projectRepoOwner string, projectRepoName string, parentPathSha256 string, fOpts []FormulaOptions) ([]FormulaStatus, error) {
	versionStr := versionutils.GetReleaseVersionOrExitGracefully().String()
	version := versionStr[1:]

	ctx := context.Background()
	client, err := githubutils.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	// Get version tag SHA
	// GitHub API docs: https://developer.github.com/v3/git/refs/#get-a-reference
	ref, _, err := client.Git.GetRef(ctx, projectRepoOwner, projectRepoName, "refs/tags/"+versionStr)
	if err != nil {
		return nil, err
	}

	versionSha := *ref.Object.SHA

	shas, err := getLocalBinarySha256(parentPathSha256)
	if err != nil {
		return nil, err
	}

	status := make([]FormulaStatus, len(fOpts))

	for i, fOpt := range fOpts {
		status[i].Name = fOpt.Name
		status[i].Updated = false

		branchName := fOpt.FormulaName + "-" + version
		commitString := fOpt.FormulaName + ": update " + version

		if fOpt.PRRepoName == fOpt.RepoName && fOpt.PRRepoOwner == fOpt.RepoOwner {
			err = updateAndPushAllRemote(client, ctx, version, versionSha, branchName, commitString, shas, &fOpt)
		} else {
			// GitHub APIs do NOT have a way to git pull --ff <remote repo>, so need to clone implementation repo locally
			// and pull remote updates.
			err = updateAndPushLocalClone(version, versionSha, branchName, commitString, shas, &fOpt, true)
		}
		if err != nil {
			if err == ErrAlreadyUpdated {
				status[i].Updated = true
			}
			status[i].Err = err
			continue
		}

		// Do NOT create PR when dryRun
		if !fOpt.dryRun {
			prRepoOwner := fOpt.PRRepoOwner
			prRepoName := fOpt.PRRepoName
			prHead := branchName // For same repo PR case

			if prRepoOwner == "" {
				prRepoOwner = fOpt.RepoOwner
				prRepoName = fOpt.RepoName
			} else if prRepoOwner != fOpt.RepoOwner {
				// For cross-repo PR, prHead should be in format of "<change repo owner>:branch"
				prHead = fOpt.RepoOwner + ":" + branchName
			}

			base := fOpt.PRBranch
			if base == "" {
				base = PRBaseBranchDefault
			}

			// Create GitHub Pull Request
			// GitHub API docs: https://developer.github.com/v3/pulls/#create-a-pull-request
			_, _, err = client.PullRequests.Create(ctx, prRepoOwner, prRepoName, &github.NewPullRequest{
				Title:               github.String(commitString),
				Head:                github.String(prHead),
				Base:                github.String(base),
				Body:                github.String(fOpt.PRDescription),
				MaintainerCanModify: github.Bool(true),
			})
			if err != nil {
				status[i].Err = err
				continue
			}
		}

		status[i].Updated = true
	}

	return status, nil
}

var (
	ErrAlreadyUpdated = errors.New("pkgmgmtutils: formula already updated")
)

func updateFormula(byt []byte, version string, versionSha string, shas *sha256Outputs, fOpt *FormulaOptions) ([]byte, error) {
	// Update Version
	if fOpt.VersionRegex != "" {
		re := regexp.MustCompile(fOpt.VersionRegex)

		// Check if formula has already been updated
		if matches := re.FindSubmatch(byt); len(matches) > 1 && bytes.Compare(matches[1], []byte(version)) == 0 {
			return byt, ErrAlreadyUpdated
		}

		byt = replaceSubmatch(byt, []byte(version), re)
	}

	// Update Version SHA (git tag sha)
	if fOpt.VersionShaRegex != "" {
		byt = replaceSubmatch(byt, []byte(versionSha), regexp.MustCompile(fOpt.VersionShaRegex))
	}

	// Update Mac SHA256
	if fOpt.DarwinShaRegex != "" {
		byt = replaceSubmatch(byt, shas.darwinSha, regexp.MustCompile(fOpt.DarwinShaRegex))
	}

	// Update Linux SHA256
	if fOpt.LinuxShaRegex != "" {
		byt = replaceSubmatch(byt, shas.linuxSha, regexp.MustCompile(fOpt.LinuxShaRegex))
	}

	// Update Windows SHA256
	if fOpt.WindowsShaRegex != "" {
		byt = replaceSubmatch(byt, shas.windowsSha, regexp.MustCompile(fOpt.WindowsShaRegex))
	}

	return byt, nil
}

func updateAndPushLocalClone(version string, versionSha string, branchName string, commitString string, shas *sha256Outputs, fOpt *FormulaOptions, mergeRemote bool) error {
	// create temp dir for local git clone
	dirTemp, err := ioutil.TempDir("", fOpt.RepoName)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dirTemp) // Cleanup local clone when done

	// git Clone Github repo
	repo, err := git.PlainClone(dirTemp, false, &git.CloneOptions{
		URL: "https://github.com/" + fOpt.RepoOwner + "/" + fOpt.RepoName,
	})
	if err != nil {
		return err
	}

	// Get git repo contents locally
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	if mergeRemote {
		// git Pull remote PR repo
		_, err = repo.CreateRemote(&config.RemoteConfig{
			Name: "upstream",
			URLs: []string{"https://github.com/" + fOpt.PRRepoOwner + "/" + fOpt.PRRepoName + ".git"},
		})
		if err != nil {
			return err
		}

		base := fOpt.PRBranch
		if base == "" {
			base = PRBaseBranchDefault
		}

		err = w.Pull(&git.PullOptions{
			RemoteName:    "upstream",
			ReferenceName: plumbing.NewBranchReferenceName(base),
		})
		if err != nil {
			return err
		}
	}

	// git checkout -b <branchName>
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: true,
	})
	if err != nil {
		return err
	}

	formulaPath := filepath.Join(dirTemp, fOpt.Path)

	byt, err := ioutil.ReadFile(formulaPath)
	if err != nil {
		return err
	}

	// Update formula with new version information
	byt, err = updateFormula(byt, version, versionSha, shas, fOpt)
	if err != nil {
		return err
	}

	// Write Updated file to git clone directory
	err = ioutil.WriteFile(formulaPath, byt, 0644)
	if err != nil {
		return err
	}

	// git commit --all
	_, err = w.Commit(commitString, &git.CommitOptions{
		All: true,
		Author: &object.Signature{
			Name:  fOpt.PRCommitName,
			Email: fOpt.PRCommitEmail,
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

func updateAndPushAllRemote(client *github.Client, ctx context.Context, version string, versionSha string, branchName string, commitString string, shas *sha256Outputs, fOpt *FormulaOptions) error {
	// Get original Formula contents
	// GitHub API docs: https://developer.github.com/v3/repos/contents/#get-contents
	fileContent, _, _, err := client.Repositories.GetContents(ctx, fOpt.RepoOwner, fOpt.RepoName, fOpt.Path, &github.RepositoryContentGetOptions{
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
	byt, err := updateFormula([]byte(c), version, versionSha, shas, fOpt)
	if err != nil {
		return err
	}

	// get master branch reference
	// GitHub API docs: https://developer.github.com/v3/git/refs/#get-a-reference
	baseRef, _, err := client.Git.GetRef(ctx, fOpt.RepoOwner, fOpt.RepoName, "refs/heads/master")
	if err != nil {
		return err
	}

	// create new branch from master branch
	// GitHub API docs: https://developer.github.com/v3/git/refs/#create-a-reference
	_, _, err = client.Git.CreateRef(ctx, fOpt.RepoOwner, fOpt.RepoName, &github.Reference{
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
	_, _, err = client.Repositories.UpdateFile(ctx, fOpt.RepoOwner, fOpt.RepoName, fOpt.Path, &github.RepositoryContentFileOptions{
		Message: github.String(commitString),
		Content: byt,
		SHA:     fileContent.SHA,
		Branch:  github.String(branchName),
		Committer: &github.CommitAuthor{
			Name:  github.String(fOpt.PRCommitName),
			Email: github.String(fOpt.PRCommitEmail),
			Date:  &now,
		},
	})
	if err != nil {
		return err
	}

	return nil
}
