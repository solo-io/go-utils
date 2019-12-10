package main

import (
	"context"
	"flag"
	"io/ioutil"

	"github.com/solo-io/go-utils/errors"

	"github.com/ghodss/yaml"
	"github.com/solo-io/go-utils/cmd/cloudbuildpreparation/pkg/api"

	"github.com/solo-io/go-utils/tarutils"
	"github.com/spf13/afero"

	"github.com/solo-io/go-utils/githubutils"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	err := run(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("unable to complete cloud build preparation", zap.Error(err))
	}

}

const (
	buildSpecFileFlagName = "spec"
)

func run(ctx context.Context) error {

	buildFile := ""
	flag.StringVar(&buildFile, buildSpecFileFlagName, "", "filename of build preparation specification")
	flag.Parse()

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
	//err = fs.MkdirAll("tmp", 0755)
	//if err != nil {
	//	return err
	//}
	file, err := ioutil.TempFile("", "new-file")
	if err != nil {
		return err
	}

	if err := githubutils.DownloadRepoArchive(ctx, githubClient, file, spec.GithubRepo.Owner, spec.GithubRepo.Repo, spec.GithubRepo.Sha); err != nil {
		contextutils.LoggerFrom(ctx).Warnw("could not download repo", zap.Error(err))
		return err
	}

	err = tarutils.Untar(spec.GithubRepo.OutputDir, file.Name(), fs)
	if err != nil {
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
