package botutils

import (
	"context"
	"fmt"
	"net/http"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"

	"github.com/google/go-github/github"
	v1 "github.com/solo-io/build/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/gcloudutils"
	"github.com/solo-io/go-utils/protoutils"
	"github.com/solo-io/go-utils/stringutils"
)

type FetchedConfig struct {
	Owner              string
	Repo               string
	Ref                string
	Config             *v1.BuildConfig
	SlackNotifications *SlackNotifications
	Error              error
}

func (fc FetchedConfig) Missing() bool {
	return fc.Config == nil && fc.Error == nil
}

func (fc FetchedConfig) Valid() bool {
	return fc.Config != nil && fc.Error == nil
}

func (fc FetchedConfig) Invalid() bool {
	return fc.Error != nil
}

func (fc FetchedConfig) String() string {
	return fmt.Sprintf("%s/%s ref=%s", fc.Owner, fc.Repo, fc.Ref)
}

type ConfigFetcher struct {
	configPath  string
	appConfig   *ApplicationConfig
	buildConfig *v1.BuildConfig
}

func NewConfigFetcher(configPath string, appConfig *ApplicationConfig, buildConfig *v1.BuildConfig) *ConfigFetcher {
	return &ConfigFetcher{configPath: configPath, appConfig: appConfig, buildConfig: buildConfig}
}

var (
	InvalidProjectError = errors.New("invalid project id")
	InvalidProjectId    = func(projectId string) error {
		return errors.Wrapf(InvalidProjectError, projectId)
	}
	projectName = ""
)

// ConfigForPR fetches the configuration for a PR. It returns an error
// only if the existence of the configuration file could not be determined. If the file
// does not exist or is invalid, the returned error is nil and the appropriate
// fields are set on the FetchedConfig.
func (cf *ConfigFetcher) ConfigForPR(ctx context.Context, client *github.Client, pr *github.PullRequest) (*FetchedConfig, error) {
	fc := &FetchedConfig{
		Owner: pr.GetBase().GetRepo().GetOwner().GetLogin(),
		Repo:  pr.GetBase().GetRepo().GetName(),
		Ref:   pr.GetBase().GetRef(),
	}

	return cf.configForCommon(ctx, client, fc)
}

func (cf *ConfigFetcher) ConfigForRelease(ctx context.Context, client *github.Client, release *github.ReleaseEvent) (*FetchedConfig, error) {
	fc := &FetchedConfig{
		Owner: release.GetRepo().GetOwner().GetLogin(),
		Repo:  release.GetRepo().GetName(),
		Ref:   release.GetRepo().GetDefaultBranch(),
	}

	return cf.configForCommon(ctx, client, fc)
}

func (cf *ConfigFetcher) ConfigForIssueComment(ctx context.Context, client *github.Client, event *github.IssueCommentEvent) (*FetchedConfig, error) {
	fc := &FetchedConfig{
		Owner: event.GetRepo().GetOwner().GetLogin(),
		Repo:  event.GetRepo().GetName(),
		Ref:   event.GetRepo().GetDefaultBranch(),
	}

	return cf.configForCommon(ctx, client, fc)
}

func (cf *ConfigFetcher) ConfigForCommitComment(ctx context.Context, client *github.Client, event *github.CommitCommentEvent) (*FetchedConfig, error) {
	fc := &FetchedConfig{
		Owner: event.GetRepo().GetOwner().GetLogin(),
		Repo:  event.GetRepo().GetName(),
		Ref:   event.GetRepo().GetDefaultBranch(),
	}

	return cf.configForCommon(ctx, client, fc)
}

func (cf *ConfigFetcher) configForCommon(ctx context.Context, client *github.Client, fc *FetchedConfig) (*FetchedConfig, error) {
	projectId := gcloudutils.GetProjectIdFromBuildConfig(cf.buildConfig)
	if !stringutils.ContainsString(projectId, cf.appConfig.GcloudProjects) {
		return nil, InvalidProjectId(projectId)
	}

	fc.SlackNotifications = &cf.appConfig.SlackNotifications

	bytes, err := cf.fetchConfigContents(ctx, client, fc.Owner, fc.Repo, fc.Ref, cf.configPath)
	if err == nil && bytes != nil {
		config, err := cf.unmarshalConfig(bytes)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("repository configuration is invalid", zap.Error(err))
		} else {
			fc.Config = config
			return fc, nil
		}
	}

	if cf.buildConfig != nil {
		contextutils.LoggerFrom(ctx).Infow("default app config used as a fallback")
		fc.Config = cf.buildConfig
		return fc, nil
	}

	fc.Error = errors.Errorf("Unable to find valid configuration")
	return fc, nil
}

// fetchConfigContents returns a nil slice if there is no configuration file
func (cf *ConfigFetcher) fetchConfigContents(ctx context.Context, client *github.Client, owner, repo, ref, configPath string) ([]byte, error) {
	opts := &github.RepositoryContentGetOptions{
		Ref: ref,
	}

	file, _, _, err := client.Repositories.GetContents(ctx, owner, repo, configPath, opts)
	if err != nil {
		if rerr, ok := err.(*github.ErrorResponse); ok && rerr.Response.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed to fetch content of %q", configPath)
	}

	// file will be nil if the ref contains a directory at the expected file path
	if file == nil {
		return nil, nil
	}

	content, err := file.GetContent()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode content of %q", configPath)
	}

	return []byte(content), nil
}

func (cf *ConfigFetcher) unmarshalConfig(bytes []byte) (*v1.BuildConfig, error) {
	var config v1.BuildConfig
	if err := protoutils.UnmarshalYaml(bytes, &config); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal configuration")
	}

	return &config, nil
}
