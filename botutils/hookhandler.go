package botutils

import (
	"context"
	"encoding/json"

	"github.com/google/go-github/github"
	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/palantir/go-githubapp/githubapp"
)

type githubHookHandler struct {
	ctx           context.Context
	clientCreator githubapp.ClientCreator
	configFetcher *ConfigFetcher
	registry      *Registry
}

func NewGithubHookHandler(ctx context.Context, clientCreator githubapp.ClientCreator, configFetcher *ConfigFetcher) *githubHookHandler {
	return &githubHookHandler{ctx: ctx, clientCreator: clientCreator, configFetcher: configFetcher, registry: &Registry{}}
}

func (h *githubHookHandler) RegisterPlugin(plugin Plugin) {
	h.registry.RegisterPlugin(plugin)
}

func (h *githubHookHandler) Handles() []string {
	return []string{PrType, PrReviewType, IssueCommentType, CommitCommentType, ReleaseType}
}

func (h *githubHookHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	// TODO: figure out which context to use
	ctx = h.ctx
	switch eventType {
	case PrType:
		return h.HandlePR(ctx, eventType, deliveryID, payload)
	case PrReviewType:
		return h.HandlePR(ctx, eventType, deliveryID, payload)
	case IssueCommentType:
		return h.HandleIssueComment(ctx, eventType, deliveryID, payload)
	case CommitCommentType:
		return h.HandleCommitComment(ctx, eventType, deliveryID, payload)
	case ReleaseType:
		return h.HandleRelease(ctx, eventType, deliveryID, payload)
	default:
		return nil
	}
}

func (h *githubHookHandler) getInstallationClient(ctx context.Context, installationId int64) (*github.Client, error) {
	client, err := h.clientCreator.NewInstallationClient(installationId)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error creating github client for installation",
			zap.Error(err),
			zap.Int64("installationId", installationId))
		return nil, err
	}
	contextutils.LoggerFrom(ctx).Infow("Created client", zap.Int64("installationId", installationId))
	return client, nil
}

func (h *githubHookHandler) configForPR(ctx context.Context, client *github.Client, pullRequest *github.PullRequest) (*FetchedConfig, error) {
	cfg, err := h.configFetcher.ConfigForPR(ctx, client, pullRequest)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error fetching config for PR", zap.Error(err))
		return nil, err
	}
	return cfg, nil
}

func (h *githubHookHandler) HandlePR(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var event github.PullRequestEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return errors.Wrap(err, "failed to parse pr event payload")
	}
	installationId := githubapp.GetInstallationIDFromEvent(&event)
	client, err := h.getInstallationClient(ctx, installationId)
	if err != nil {
		return err
	}
	cfg, err := h.configForPR(ctx, client, event.PullRequest)
	if err != nil {
		return err
	}
	go h.registry.CallPrPlugins(ctx, client, cfg, &event)
	return nil
}

func (h *githubHookHandler) HandlePrReview(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var event github.PullRequestReviewEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return errors.Wrap(err, "failed to parse pr review event payload")
	}
	installationId := githubapp.GetInstallationIDFromEvent(&event)
	client, err := h.getInstallationClient(ctx, installationId)
	if err != nil {
		return err
	}
	cfg, err := h.configForPR(ctx, client, event.PullRequest)
	if err != nil {
		return err
	}
	go h.registry.PullRequestReviewPlugins(ctx, client, cfg, &event)
	return nil
}

func (h *githubHookHandler) HandleIssueComment(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var event github.IssueCommentEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return errors.Wrap(err, "failed to parse issue comment event payload")
	}
	installationId := githubapp.GetInstallationIDFromEvent(&event)
	client, err := h.getInstallationClient(ctx, installationId)
	if err != nil {
		return err
	}
	cfg, err := h.configFetcher.ConfigForIssueComment(ctx, client, &event)
	if err != nil {
		return err
	}
	go h.registry.CallIssueCommentPlugins(ctx, client, cfg, &event)
	return nil
}

func (h *githubHookHandler) HandleCommitComment(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var event github.CommitCommentEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return errors.Wrap(err, "failed to parse commit comment event payload")
	}
	installationId := githubapp.GetInstallationIDFromEvent(&event)
	client, err := h.getInstallationClient(ctx, installationId)
	if err != nil {
		return err
	}
	cfg, err := h.configFetcher.ConfigForCommitComment(ctx, client, &event)
	if err != nil {
		return err
	}
	go h.registry.CallCommitCommentPlugins(ctx, client, cfg, &event)
	return nil
}

func (h *githubHookHandler) HandleRelease(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var event github.ReleaseEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return errors.Wrap(err, "failed to parse release event payload")
	}
	installationId := githubapp.GetInstallationIDFromEvent(&event)
	client, err := h.getInstallationClient(ctx, installationId)
	if err != nil {
		return err
	}
	cfg, err := h.configFetcher.ConfigForRelease(ctx, client, &event)
	if err != nil {
		return err
	}
	go h.registry.CallReleasePlugins(ctx, client, cfg, &event)
	return nil
}
