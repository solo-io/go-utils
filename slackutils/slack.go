package slackutils

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
)

var _ SlackClient = new(slackClient)

type SlackNotifications struct {
	DefaultUrl string            `yaml:"default_url" json:"defaultUrl"`
	RepoUrls   map[string]string `yaml:"repo_urls" json:"repoUrls"`
}

type SlackClient interface {
	// Use repo-specific channel, if exists
	NotifyForRepo(ctx context.Context, repo, message string)
	// Use default channel
	Notify(ctx context.Context, message string)
}

func NewSlackClient(notifications *SlackNotifications) *slackClient {
	return NewSlackClientForHttpClient(&DefaultHttpClient{}, notifications)
}

func NewSlackClientForHttpClient(httpClient HttpClient, notifications *SlackNotifications) *slackClient {
	return &slackClient{
		httpClient:    httpClient,
		notifications: notifications,
	}
}

type slackClient struct {
	httpClient    HttpClient
	notifications *SlackNotifications
}

func (s *slackClient) getSlackUrl(repo string) string {
	if s.notifications == nil {
		return ""
	}
	if repo == "" || s.notifications.RepoUrls == nil {
		return s.notifications.DefaultUrl
	}
	repoUrl, ok := s.notifications.RepoUrls[repo]
	if ok {
		return repoUrl
	}
	return s.notifications.DefaultUrl
}

func (s *slackClient) Notify(ctx context.Context, message string) {
	s.NotifyForRepo(ctx, "", message)
}

func (s *slackClient) NotifyForRepo(ctx context.Context, repo, message string) {
	slackUrl := s.getSlackUrl(repo)
	if slackUrl == "" {
		contextutils.LoggerFrom(ctx).Warnw("Requested notifying slack, but no URL",
			zap.String("repo", repo),
			zap.Any("events", s.notifications))
		return
	}
	s.httpClient.PostJsonContent(ctx, message, slackUrl)
}
