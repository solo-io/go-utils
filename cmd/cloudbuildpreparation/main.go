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
