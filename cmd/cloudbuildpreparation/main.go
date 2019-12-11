package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/solo-io/go-utils/errors"

	"github.com/ghodss/yaml"
	"github.com/solo-io/go-utils/cmd/cloudbuildpreparation/pkg/api"

	"github.com/solo-io/go-utils/tarutils"
	"github.com/spf13/afero"

	"github.com/solo-io/go-utils/githubutils"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
)

const shortDescriptiveName = "cloud build preparation"

func main() {
	ctx := context.Background()
	contextutils.LoggerFrom(ctx).Infow("starting")
	err := run(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("unable to complete cloud build preparation", zap.Error(err))
	}
	contextutils.LoggerFrom(ctx).Infow("completed successfully")
}

const (
	buildSpecFileFlagName = "spec"
)

func run(ctx context.Context) error {

	buildFile := ""
	flag.StringVar(&buildFile, buildSpecFileFlagName, "", "filename of build preparation specification")
	flag.Parse()
	contextutils.LoggerFrom(ctx).Infow("reading config from file", zap.Any("filename", buildFile))

	spec, err := ingestBuildSpec(buildFile)
	if err != nil {
		return err
	}

	// This utility expects GITHUB_TOKEN to exist in the environment
	githubClient, err := githubutils.GetClient(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Warnw("could not get github client", zap.Error(err))
		return err
	}

	fs := afero.NewOsFs()
	file, err := ioutil.TempFile("", "new-file")
	if err != nil {
		return err
	}
	tempDir, err := ioutil.TempDir("", "")
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
	if err := renameOutputFile(tempDir, spec.GithubRepo); err != nil {
		return err
	}
	// cleanup
	if err := os.RemoveAll(tempDir); err != nil {
		return err
	}
	return nil
}

func ingestBuildSpec(filename string) (*api.BuildPreparation, error) {
	if filename == "" {
		return nil, errors.Errorf("must provide a spec filename with flag --%v", buildSpecFileFlagName)
	}
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	spec := &api.BuildPreparation{}
	if err = yaml.Unmarshal(content, spec); err != nil {
		return nil, err
	}
	if err = validateSpec(spec); err != nil {
		return nil, err
	}
	return spec, nil
}

func validateSpec(spec *api.BuildPreparation) error {
	if spec == nil {
		return errors.New("invalid spec: spec is empty")
	}
	return nil
}

func renameOutputFile(tempDir string, gitHubSpec api.GithubRepo) error {
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
	if err := os.MkdirAll(parentDir, os.ModePerm); err != nil {
		return err
	}
	newName := filepath.Join(parentDir, gitHubSpec.Repo)
	return os.Rename(filepath.Join(tempDir, oldName), newName)
}
