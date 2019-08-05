package builders

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/solo-io/go-utils/gcloudutils"
	"net/url"
	"time"

	"github.com/avast/retry-go"

	"github.com/rs/zerolog"
	"google.golang.org/api/cloudbuild/v1"
)

type subscription = string

const (
	PROJECT_ID  subscription = "PROJECT_ID"
	BUILD_ID                 = "BUILD_ID"
	COMMIT_SHA               = "COMMIT_SHA"
	SHORT_SHA                = "SHORT_SHA"
	REPO_NAME                = "REPO_NAME"
	BRANCH_NAME              = "BRANCH_NAME"
	TAG_NAME                 = "TAG_NAME"
	REVISION_ID              = "REVISION_ID"

	soloLogBucket = "solo-public-build-logs"
)

type Builder interface {
	InitBuildWithSha(ctx context.Context, builderCtx ShaBuildContext) (*cloudbuild.Build, error)
	InitBuildWithTag(ctx context.Context, builderCtx TagBuildContext) (*cloudbuild.Build, error)
}

type BuildInfo struct {
	LogUrl string
}

func ToEnv(s subscription) string {
	return fmt.Sprintf("$%s", s)
}

type OperationBuildMetadata struct {
	Type  string            `json:"@type"`
	Build *cloudbuild.Build `json:"build"`
}

func startBuild(ctx context.Context, builderCtx BuildContext, cbm *cloudbuild.Build) (*BuildInfo, error) {
	logger := zerolog.Ctx(ctx)
	repo, buildClient := builderCtx.Repo(), builderCtx.Service()
	setLogBucket(ctx, builderCtx, cbm)
	createCall := buildClient.Projects.Builds.Create(builderCtx.ProjectId(), cbm)
	var op *cloudbuild.Operation

	// simple retry logic if can't find source
	err := retry.Do(
		func() error {
			var err error
			op, err = createCall.Context(ctx).Do()
			if err != nil {
				return err
			}
			return nil
		},
		retry.DelayType(retry.BackOffDelay),
		retry.Attempts(6),
		retry.Delay(2*time.Second),
		retry.RetryIf(func(e error) bool {
			return gcloudutils.IsMissingSourceError(e.Error())
		}),
	)
	if err != nil {
		return nil, err
	}

	logger.Info().Msgf("successfully started build for %s", repo)

	byt, err := op.Metadata.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var buildMetadata OperationBuildMetadata
	if err := json.Unmarshal(byt, &buildMetadata); err != nil {
		return nil, err
	}

	bi := BuildInfo{
		LogUrl: buildMetadata.Build.LogUrl,
	}

	if gcloudutils.IsPublic(ctx, builderCtx.ProjectId(), builderCtx.Repo()) {
		bi.LogUrl = fmt.Sprintf("https://storage.googleapis.com/%s/logs.html?buildid=%s", soloLogBucket, url.QueryEscape(buildMetadata.Build.Id))
	}

	return &bi, nil
}

func StartBuildWithTag(ctx context.Context, builderCtx TagBuildContext) (*BuildInfo, error) {
	storageBuilder, err := NewStorageBuilder(ctx, builderCtx.ProjectId())
	if err != nil {
		return nil, err
	}
	cbm, err := storageBuilder.InitBuildWithTag(ctx, builderCtx)
	if err != nil {
		return nil, err
	}

	tags := gcloudutils.InitializeTags(cbm.Tags)
	tags = tags.AddReleaseTag(builderCtx.Tag())
	tags = tags.AddRepoTag(builderCtx.Repo())
	tags = tags.AddInstallationIdTag(builderCtx.InstallationId())
	cbm.Tags = tags

	return startBuild(ctx, builderCtx, cbm)
}

func StartBuildWithSha(ctx context.Context, builderCtx ShaBuildContext) (*BuildInfo, error) {
	storageBuilder, err := NewStorageBuilder(ctx, builderCtx.ProjectId())
	if err != nil {
		return nil, err
	}
	cbm, err := storageBuilder.InitBuildWithSha(ctx, builderCtx)
	if err != nil {
		return nil, err
	}

	tags := gcloudutils.InitializeTags(cbm.Tags)
	tags = tags.AddInstallationIdTag(builderCtx.InstallationId())
	tags = tags.AddShaTag(builderCtx.Sha())
	tags = tags.AddRepoTag(builderCtx.Repo())
	cbm.Tags = tags

	return startBuild(ctx, builderCtx, cbm)
}

func DescribeBuild(ctx context.Context, builderCtx BuildContext, id string) (*cloudbuild.Build, error) {
	buildClient := builderCtx.Service()
	projectId := builderCtx.ProjectId()

	getCall := buildClient.Projects.Builds.Get(projectId, id)
	result, err := getCall.Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func setLogBucket(ctx context.Context, builderCtx BuildContext, cbm *cloudbuild.Build) {
	logger := zerolog.Ctx(ctx)
	if !gcloudutils.IsPublic(ctx, builderCtx.ProjectId(), builderCtx.Repo()) {
		logger.Debug().Msg("project is not public")
		return
	}
	cbm.LogsBucket = soloLogBucket
	logger.Debug().Str("LogsBucket", soloLogBucket).Msg("setting logs bucket")
}
