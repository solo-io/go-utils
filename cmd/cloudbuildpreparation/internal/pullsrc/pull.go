package pullsrc

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/go-github/github"
	"github.com/solo-io/go-utils/cmd/cloudbuildpreparation/pkg/api"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/tarutils"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

func GitCloneSourceCode(ctx context.Context, spec *api.BuildPreparation) error {
	if err := mkdir(ctx, spec.GithubRepo.OutputDir); err != nil {
		return err
	}
	cloneUrl := fmt.Sprintf("https://www.github.com/%v/%v.git", spec.GithubRepo.Owner, spec.GithubRepo.Repo)
	cmd := commonGitCmd(spec, "clone", cloneUrl)
	cmd.Dir = spec.GithubRepo.OutputDir
	return cmd.Run()
}
func GitCheckoutSha(ctx context.Context, spec *api.BuildPreparation) error {
	cmd := commonGitCmd(spec, "checkout", spec.GithubRepo.Sha)
	cmd.Dir = getPathToRepoDir(spec)
	return cmd.Run()
}
func GitDescribeTagsDirty(ctx context.Context, spec *api.BuildPreparation) error {
	cmd := commonGitCmd(spec, "describe", "--tags", "--dirty", "--always")
	// --tags: shows the most recent tag, sha content, and how many commits have been made since the tag
	// --dirty: appends "-dirty" if tracked files have been edited (but not if new untracked files have been added)
	// --always: provides sha content even when there has been no tag
	// cmd: git describe --tags --always --dirty
	// output when no tags ever defined, clean repo: 123abcd
	// output when no tags ever defined, dirty repo: 123abcd-dirty
	// output when tags have been defined, clean repo: v1.2.7-8-123abcd
	// output when tags have been defined, dirty repo: v1.2.7-8-123abcd-dirty
	cmd.Dir = getPathToRepoDir(spec)
	return cmd.Run()
}
func getPathToRepoDir(spec *api.BuildPreparation) string {
	if spec.GithubRepo.OutputDir != "" {
		return fmt.Sprintf("%v/%v", spec.GithubRepo.OutputDir, spec.GithubRepo.Repo)
	}
	return spec.GithubRepo.Repo
}
func commonGitCmd(spec *api.BuildPreparation, args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func PullSourceCode(ctx context.Context, githubClient *github.Client, spec *api.BuildPreparation) error {
	if githubClient == nil {
		var err error
		// This utility expects GITHUB_TOKEN to exist in the environment
		githubClient, err = githubutils.GetClient(ctx)
		if err != nil {
			contextutils.LoggerFrom(ctx).Warnw("could not get github client", zap.Error(err))
			return err
		}
	}

	fs := afero.NewOsFs()
	file, err := ioutil.TempFile("", "new-file")
	if err != nil {
		return err
	}

	// need to use the designated output dir here otherwise this script will fail when
	// we rename the github-generated unarchived directory if the output dir is on a different device.
	// For example, if this script is run in docker and references a mounted
	if err := mkdir(ctx, spec.GithubRepo.OutputDir); err != nil {
		return err
	}
	tempDir, err := ioutil.TempDir(spec.GithubRepo.OutputDir, "")
	if err != nil {
		return err
	}

	if err := githubutils.DownloadRepoArchive(ctx, githubClient, file, spec.GithubRepo.Owner, spec.GithubRepo.Repo, spec.GithubRepo.Sha); err != nil {
		contextutils.LoggerFrom(ctx).Warnw("could not download repo", zap.Error(err))
		return err
	}
	err = tarutils.Untar(tempDir, file.Name(), fs)
	if err != nil {
		return err
	}

	// GitHub's archives include a portion of the sha. This is nice but is harder to predict or read by scripts
	// Rename to the repo name
	if err := renameOutputFile(ctx, tempDir, spec.GithubRepo); err != nil {
		return err
	}

	// cleanup
	if err := os.RemoveAll(tempDir); err != nil {
		return err
	}
	return nil
}

func renameOutputFile(ctx context.Context, tempDir string, gitHubSpec api.GithubRepo) error {
	// find the file that was written
	fileInfo, err := ioutil.ReadDir(tempDir)
	if err != nil {
		return err
	}
	if len(fileInfo) != 1 {
		return errors.Errorf("expected a single entry in temp dir, found %v", len(fileInfo))
	}
	oldName := fileInfo[0].Name()

	// rename it
	parentDir := filepath.Join(gitHubSpec.OutputDir, gitHubSpec.Owner)
	if err := mkdir(ctx, parentDir); err != nil {
		return err
	}
	newName := filepath.Join(parentDir, gitHubSpec.Repo)
	return os.Rename(filepath.Join(tempDir, oldName), newName)
}

func mkdir(ctx context.Context, dir string) error {
	if dir == "" {
		return nil
	}
	contextutils.LoggerFrom(ctx).Infow("creating dir",
		zap.Any("dirname", dir))
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		contextutils.LoggerFrom(ctx).Warnw("error while creating dir",
			zap.Error(err),
			zap.Any("dirname", dir))
		return err
	}
	return nil
}
