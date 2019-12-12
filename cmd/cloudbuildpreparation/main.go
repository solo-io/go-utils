package main

import (
	"context"
	"flag"
	"io/ioutil"

	"github.com/solo-io/go-utils/errors"

	"github.com/ghodss/yaml"
	"github.com/solo-io/go-utils/cmd/cloudbuildpreparation/internal/pullsrc"
	"github.com/solo-io/go-utils/cmd/cloudbuildpreparation/pkg/api"

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
	repoShaFlagName       = "repo-sha"
	repoOwnerFlagName     = "repo-owner"
	repoNameFlagName      = "repo-name"
	repoOutputDirFlagName = "repo-output-dir"
)

func run(ctx context.Context) error {

	spec, err := gatherCliInput(ctx)
	if err != nil {
		return err
	}

	if err := pullsrc.GitCloneSourceCode(ctx, spec); err != nil {
		return err
	}
	if err := pullsrc.GitCheckoutSha(ctx, spec); err != nil {
		return err
	}
	if err := pullsrc.GitDescribeTagsDirty(ctx, spec); err != nil {
		return err
	}
	//return pullsrc.PullSourceCode(ctx, githubClient, spec)
	return nil
}

func gatherCliInput(ctx context.Context) (*api.BuildPreparation, error) {
	// flags
	buildFile := ""
	cliFlagRepoContent := &api.GithubRepo{}
	flag.StringVar(&buildFile, buildSpecFileFlagName, "", "optional, filename of build preparation specification - if provided, repo-* flags should be skipped")
	flag.StringVar(&cliFlagRepoContent.Sha, repoShaFlagName, "", "sha (or tag or branch) to checkout")
	flag.StringVar(&cliFlagRepoContent.Owner, repoOwnerFlagName, "", "repo owner")
	flag.StringVar(&cliFlagRepoContent.Repo, repoNameFlagName, "", "repo name")
	flag.StringVar(&cliFlagRepoContent.OutputDir, repoOutputDirFlagName, "", "directory into which to clone repo")
	flag.Parse()

	if err := validateCliInput(buildFile, cliFlagRepoContent); err != nil {
		return nil, err
	}

	// translate flags to spec
	if buildFile != "" {
		contextutils.LoggerFrom(ctx).Infow("reading config from file", zap.Any("filename", buildFile))
		return ingestBuildSpec(buildFile, cliFlagRepoContent)
	} else {
		return &api.BuildPreparation{
			GithubRepo: *cliFlagRepoContent,
		}, nil
	}
}

func ingestBuildSpec(filename string, cliFlagRepoContent *api.GithubRepo) (*api.BuildPreparation, error) {
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
	if err = validateSpec(spec, cliFlagRepoContent); err != nil {
		return nil, err
	}
	return spec, nil
}

func validateCliInput(filename string, cliFlagRepoContent *api.GithubRepo) error {
	if filename == "" {
		if cliFlagRepoContent.Repo == "" {
			return errors.Errorf("must provide a repo name with --%v", repoNameFlagName)
		}
		if cliFlagRepoContent.Owner == "" {
			return errors.Errorf("must provide a repo owner with --%v", repoOwnerFlagName)
		}
		if cliFlagRepoContent.Sha == "" {
			return errors.Errorf("must provide a repo sha with --%v", repoShaFlagName)
		}
	}
	return nil
}
func validateSpec(spec *api.BuildPreparation, cliFlagRepoContent *api.GithubRepo) error {
	if spec == nil {
		return errors.New("invalid spec: spec is empty")
	}
	return nil
}
