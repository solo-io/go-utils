package builders

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/cloudbuild/v1"
)

type SourceBuilder struct{}

var _ Builder = new(SourceBuilder)

func (sb *SourceBuilder) InitBuildWithSha(ctx context.Context, builderCtx ShaBuildContext) (*cloudbuild.Build, error) {
	owner, repo := builderCtx.Owner(), builderCtx.Repo()
	cbm, content, err := unmarshalCloudbuild(ctx, builderCtx, builderCtx.Sha())
	if err != nil {
		return nil, err
	}

	cbm.Source = &cloudbuild.Source{
		RepoSource: &cloudbuild.RepoSource{
			RepoName:  fmt.Sprintf("github_%s_%s", owner, repo),
			CommitSha: builderCtx.Sha(),
		},
	}

	if strings.Contains(content, ToEnv(COMMIT_SHA)) {
		cbm.Substitutions = map[string]string{
			COMMIT_SHA: builderCtx.Sha(),
		}
	}
	cbm.Substitutions = addDefaultSubsitutions(content, cbm)
	return cbm, nil
}

func (sb *SourceBuilder) InitBuildWithTag(ctx context.Context, builderCtx TagBuildContext) (*cloudbuild.Build, error) {
	owner, repo := builderCtx.Owner(), builderCtx.Repo()
	cbm, content, err := unmarshalCloudbuild(ctx, builderCtx, builderCtx.Tag())
	if err != nil {
		return nil, err
	}

	cbm.Source = &cloudbuild.Source{
		RepoSource: &cloudbuild.RepoSource{
			RepoName: fmt.Sprintf("github_%s_%s", owner, repo),
			TagName:  builderCtx.Tag(),
		},
	}

	cbm.Substitutions = make(map[string]string)

	if strings.Contains(content, ToEnv(TAG_NAME)) {
		cbm.Substitutions[TAG_NAME] = builderCtx.Tag()
	}
	if strings.Contains(content, ToEnv(COMMIT_SHA)) {
		cbm.Substitutions[COMMIT_SHA] = builderCtx.Sha()
	}

	cbm.Substitutions = addDefaultSubsitutions(content, cbm)
	return cbm, nil
}
