package botutils

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"

	"github.com/google/go-github/github"
)

type Registry struct {
	prplugins      []PullRequestHandler
	prrplugins     []PullRequestReviewHandler
	icplugins      []IssueCommentHandler
	ccplugins      []CommitCommentHandler
	releaseplugins []ReleaseHandler
}

func (r *Registry) RegisterPlugin(p Plugin) {
	if plugin, ok := p.(PullRequestHandler); ok {
		r.prplugins = append(r.prplugins, plugin)
	}
	if plugin, ok := p.(PullRequestReviewHandler); ok {
		r.prrplugins = append(r.prrplugins, plugin)
	}
	if plugin, ok := p.(IssueCommentHandler); ok {
		r.icplugins = append(r.icplugins, plugin)
	}
	if plugin, ok := p.(CommitCommentHandler); ok {
		r.ccplugins = append(r.ccplugins, plugin)
	}
	if plugin, ok := p.(ReleaseHandler); ok {
		r.releaseplugins = append(r.releaseplugins, plugin)
	}
}

func (r *Registry) CallPrPlugins(ctx context.Context, client *github.Client, fetchedConfig *FetchedConfig, event *github.PullRequestEvent) {
	for _, pr := range r.prplugins {
		err := pr.HandlePREvent(ctx, client, fetchedConfig, event)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("error handling PR", zap.Error(err))
		}
	}
}

func (r *Registry) PullRequestReviewPlugins(ctx context.Context, client *github.Client, fetchedConfig *FetchedConfig, event *github.PullRequestReviewEvent) {
	for _, pr := range r.prrplugins {
		err := pr.HandlePullRequestReviewEvent(ctx, client, fetchedConfig, event)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("error handling PR review", zap.Error(err))

		}
	}
}

func (r *Registry) CallIssueCommentPlugins(ctx context.Context, client *github.Client, fetchedConfig *FetchedConfig, event *github.IssueCommentEvent) {
	for _, pr := range r.icplugins {
		err := pr.HandleIssueCommentEvent(ctx, client, fetchedConfig, event)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("error handling issue comment", zap.Error(err))
		}
	}
}

func (r *Registry) CallCommitCommentPlugins(ctx context.Context, client *github.Client, fetchedConfig *FetchedConfig, event *github.CommitCommentEvent) {
	for _, pr := range r.ccplugins {
		err := pr.HandleCommitCommentEvent(ctx, client, fetchedConfig, event)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("error handling commit comment", zap.Error(err))
		}
	}
}

func (r *Registry) CallReleasePlugins(ctx context.Context, client *github.Client, fetchedConfig *FetchedConfig, event *github.ReleaseEvent) {
	for _, pr := range r.releaseplugins {
		err := pr.HandleReleaseEvent(ctx, client, fetchedConfig, event)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("error handling release", zap.Error(err))
		}
	}
}
