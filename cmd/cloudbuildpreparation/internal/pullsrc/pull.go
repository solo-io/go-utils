package pullsrc

import (
	"context"
	"io/ioutil"
	"os"
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

func PullSourceCode(ctx context.Context, githubClient *github.Client, spec *api.BuildPreparation) error {

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
