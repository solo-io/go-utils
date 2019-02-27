package docsutils

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/onsi/ginkgo"
	"github.com/solo-io/go-utils/changelogutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/logger"
	"github.com/spf13/afero"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	DocsRepo = "solo-docs"
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
	exists, err := afero.Exists(fs, DocsRepo)
	if err != nil {
		return err
	}
	if exists {
		return errors.Errorf("Cannot clone because %s already exists", DocsRepo)
	}
	err = gitCloneDocs()
	defer fs.RemoveAll(DocsRepo)
	if err != nil {
		return errors.Wrapf(err, "Error cloning repo")
	}

	client, err := githubutils.GetClient(ctx)
	if err != nil {
		return err
	}
	latestTag, err := githubutils.FindLatestReleaseTag(ctx, client, owner, repo)
	if err != nil {
		return err
	}
	proposedTag, err := changelogutils.GetProposedTag(fs, latestTag, "")
	if err != nil {
		return err
	}
	changelog, err := changelogutils.ComputeChangelog(fs, latestTag, proposedTag, "")
	if err != nil {
		return err
	}
	markdown := changelogutils.GenerateChangelogMarkdown(changelog)
	fmt.Printf(markdown)

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
	
	title := fmt.Sprintf("Update docs for %s %s", product, tag)
	body := fmt.Sprintf("Automatically generated docs for %s %s", product, tag)
	base := "master"
	pr := github.NewPullRequest{
		Title: &title,
		Body: &body,
		Head: &branch,
		Base: &base,
	}
	_, _, err = client.PullRequests.Create(ctx, owner, DocsRepo, &pr)
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
	return execGit("", "clone", fmt.Sprintf("https://soloio-bot:%s@github.com/solo-io/%s.git", token, DocsRepo))
}

func gitCheckoutNewBranch(branch string) error {
	return execGit(DocsRepo, "checkout", "-b", branch)
}

func gitAddAll() error {
	return execGit(DocsRepo, "add", ".")
}

func gitCommit(tag string) error {
	return execGit(DocsRepo, "commit", "-m", fmt.Sprintf("docs for %s", tag))
}

func gitDiffIsEmpty() (bool, error) {
	output, err := execGitWithOutput(DocsRepo, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return output == "", err
}

func gitPush(branch string) error {
	return execGit(DocsRepo, "push", "origin", branch)
}

func prepareCmd(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	logger.Debugf("git %v", cmd.Args)
	cmd.Env = os.Environ()
	// disable DEBUG=1 from getting through to kube
	cmd.Stdout = ginkgo.GinkgoWriter
	cmd.Stderr = ginkgo.GinkgoWriter
	cmd.Dir = dir
	return cmd
}

func execGit(dir string, args ...string) error {
	cmd := prepareCmd(dir, args...)
	return cmd.Run()
}

func execGitWithOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func replaceDirectories(product string, paths ...string) error {
	fs := afero.NewOsFs()
	for _, path := range paths {
		soloDocsPath := filepath.Join(DocsRepo, product, path)
		exists, err := afero.Exists(fs, soloDocsPath)
		if err != nil {
			return err
		}
		if exists {
			err = fs.RemoveAll(soloDocsPath)
			if err != nil {
				return err
			}
		}
		err = copyRecursive(path, soloDocsPath)
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
