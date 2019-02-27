package docsutils

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/onsi/ginkgo"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/logger"
	"github.com/spf13/afero"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

/*
Example:
CreateDocsPR("solo-io", "gloo", "v0.8.2", "gloo",
"docs/v1/github.com/solo-io/gloo",
"docs/v1/github.com/solo-io/solo-kit",
"docs/v1/gogoproto",
"docs/v1/google")
 */
func CreateDocsPR(owner, repo, tag, product string, paths ...string) error {
	ctx := context.TODO()
	fs := afero.NewOsFs()
	exists, err := afero.Exists(fs, "solo-docs")
	if err != nil {
		return err
	}
	if exists {
		return errors.Errorf("Cannot clone because solo-docs already exists")
	}
	err = gitCloneDocs()
	defer fs.RemoveAll("solo-docs")
	if err != nil {
		return errors.Wrapf(err, "Error cloning repo")
	}
	branch := repo + "-docs-" + tag
	err = gitCheckoutNewBranch(branch)
	if err != nil {
		return errors.Wrapf(err, "Error checking out branch")
	}
	err = replaceDirectories(product, paths...)
	if err != nil {
		return errors.Wrapf(err, "Error removing old docs")
	}
	err = gitAddAll()
	if err != nil {
		return errors.Wrapf(err, "Error doing git add")
	}
	empty, err := gitDiffIsEmpty()
	if err != nil {
		return errors.Wrapf(err, "Error checking for diff")
	}
	if empty {
		// no diff, exit early cause we're done
		return nil
	}

	err = gitCommit(tag)
	if err != nil {
		return errors.Wrapf(err, "Error doing git commit")
	}
	err = gitPush(branch)
	if err != nil {
		return errors.Wrapf(err, "Error pushing docs branch")
	}
	client, err := githubutils.GetClient(ctx)
	if err != nil {
		return err
	}
	title := fmt.Sprintf("Update docs for %s %s", product, tag)
	body := fmt.Sprintf("Automatically generated docs for %s %s", product, tag)
	base := "master"
	pr := github.NewPullRequest{
		Title: &title,
		Body: &body,
		Head: &branch,
		Base: &base,
	}
	_, _, err = client.PullRequests.Create(ctx, owner, repo, &pr)
	if err != nil {
		return errors.Wrapf(err, "Error creating PR")
	}
	return nil
}

func gitCloneDocs() error {
	token, err := githubutils.GetGithubToken()
	if err != nil {
		return err
	}
	return execGit(fmt.Sprintf("clone https://soloio-bot:%s@github.com/solo-io/solo-docs.git", token), "")
}

func gitCheckoutNewBranch(branch string) error {
	return execGit(fmt.Sprintf("checkout -b %s", branch), "solo-docs")
}

func gitAddAll() error {
	return execGit("add .", "solo-docs")
}

func gitCommit(tag string) error {
	return execGit(fmt.Sprintf("commit -m \"docs for %s\"", tag), "solo-docs")
}

func gitDiffIsEmpty() (bool, error) {
	output, err := execGitWithOutput("status --porcelain", "solo-docs")
	if err != nil {
		return false, err
	}
	return output == "", err
}

func gitPush(branch string) error {
	return execGit(fmt.Sprintf("push origin/%s", branch), "solo-docs")
}

func prepareCmd(argsString, dir string) *exec.Cmd {
	args := strings.Split(argsString, " ")
	cmd := exec.Command("git", args...)
	logger.Debugf("git %v", cmd.Args)
	cmd.Env = os.Environ()
	// disable DEBUG=1 from getting through to kube
	cmd.Stdout = ginkgo.GinkgoWriter
	cmd.Stderr = ginkgo.GinkgoWriter
	cmd.Dir = dir
	return cmd
}

func execGit(argsString, dir string) error {
	cmd := prepareCmd(argsString, dir)
	return cmd.Run()
}

func execGitWithOutput(argsString, dir string) (string, error) {
	cmd := prepareCmd(argsString, dir)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func replaceDirectories(product string, paths ...string) error {
	fs := afero.NewOsFs()
	for _, path := range paths {
		exists, err := afero.Exists(fs, path)
		if err != nil {
			return err
		}
		if exists {
			err = fs.RemoveAll(path)
			if err != nil {
				return err
			}
		}
		err = copyRecursive(path, filepath.Join("solo-docs", product, path))
		if err != nil {
			return err
		}
	}
	return nil
}

func copyRecursive(from, to string) error {
	cmd := exec.Command("cp", "-r", from, to)
	logger.Debugf("cp %v", cmd.Args)
	cmd.Env = os.Environ()
	// disable DEBUG=1 from getting through to kube
	cmd.Stdout = ginkgo.GinkgoWriter
	cmd.Stderr = ginkgo.GinkgoWriter
	return cmd.Run()
}
