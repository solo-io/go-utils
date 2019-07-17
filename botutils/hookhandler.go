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
	registry      *Registry
}

func NewGithubHookHandler(ctx context.Context, clientCreator githubapp.ClientCreator) *githubHookHandler {
	return &githubHookHandler{ctx: ctx, clientCreator: clientCreator, registry: &Registry{}}
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
	case IssuesType:
		return h.HandleIssues(ctx, eventType, deliveryID, payload)
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
	go h.registry.CallPrPlugins(ctx, client, &event)
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
	go h.registry.PullRequestReviewPlugins(ctx, client, &event)
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
	go h.registry.CallIssueCommentPlugins(ctx, client, &event)
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
	go h.registry.CallCommitCommentPlugins(ctx, client, &event)
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
	go h.registry.CallReleasePlugins(ctx, client, &event)
	return nil
}

func (h *githubHookHandler) HandleIssues(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var event github.IssuesEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return errors.Wrap(err, "failed to parse issues event payload")
	}
	installationId := githubapp.GetInstallationIDFromEvent(&event)
	client, err := h.getInstallationClient(ctx, installationId)
	if err != nil {
		return err
	}
	go h.registry.CallIssuesPlugins(ctx, client, &event)
	return nil
}
