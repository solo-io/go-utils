package slackutils

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
)

var _ SlackClient = new(slackClient)

type SlackNotifications struct {
	DefaultUrl string            `yaml:"default_url" json:"defaultUrl"`
	RepoUrls   map[string]string `yaml:"repo_urls" json:"repoUrls"`
}

type SlackClient interface {
	Notify(ctx context.Context, repo, message string)
}

func NewSlackClient(notifications *SlackNotifications) *slackClient {
	return &slackClient{
		notifications: notifications,
	}
}

type slackClient struct {
	notifications *SlackNotifications
}

func (s *slackClient) getSlackUrl(repo string) string {
	if s.notifications == nil || s.notifications.RepoUrls == nil {
		return ""
	}
	repoUrl := s.notifications.RepoUrls[repo]
	if repoUrl != "" {
		return repoUrl
	}
	return s.notifications.DefaultUrl
}

func (s *slackClient) Notify(ctx context.Context, repo, message string) {
	slackUrl := s.getSlackUrl(repo)
	if slackUrl == "" {
		return
	}

	type Payload struct {
		Text string `json:"text"`
	}

	data := Payload{
		Text: message,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Notifying slack failed", zap.Error(err))
		return
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.Post(slackUrl, "application/json", body)
	defer req.Body.Close()
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Notifying slack failed", zap.Error(err))
		return
	}
}
