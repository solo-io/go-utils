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
	issuesplugins  []IssuesHandler
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
	if plugin, ok := p.(IssuesHandler); ok {
		r.issuesplugins = append(r.issuesplugins, plugin)
	}
}

func (r *Registry) CallPrPlugins(ctx context.Context, client *github.Client, event *github.PullRequestEvent) {
	for _, pr := range r.prplugins {
		contextutils.LoggerFrom(ctx).Debugw("PR event",
			zap.String("owner", event.GetRepo().GetOwner().GetLogin()),
			zap.String("repo", event.GetRepo().GetName()),
			zap.Int("pr", event.GetPullRequest().GetNumber()))
		err := pr.HandlePREvent(ctx, client, event)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("error handling PR", zap.Error(err), zap.Any("event", event))
		}
	}
}

func (r *Registry) PullRequestReviewPlugins(ctx context.Context, client *github.Client, event *github.PullRequestReviewEvent) {
	for _, pr := range r.prrplugins {
		contextutils.LoggerFrom(ctx).Debugw("PR review event",
			zap.String("owner", event.GetRepo().GetOwner().GetLogin()),
			zap.String("repo", event.GetRepo().GetName()),
			zap.Int("pr", event.GetPullRequest().GetNumber()))
		err := pr.HandlePullRequestReviewEvent(ctx, client, event)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("error handling PR review", zap.Error(err), zap.Any("event", event))

		}
	}
}

func (r *Registry) CallIssueCommentPlugins(ctx context.Context, client *github.Client, event *github.IssueCommentEvent) {
	for _, pr := range r.icplugins {
		contextutils.LoggerFrom(ctx).Debugw("Issue comment",
			zap.String("owner", event.GetRepo().GetOwner().GetLogin()),
			zap.String("repo", event.GetRepo().GetName()),
			zap.Int("issue", event.GetIssue().GetNumber()),
			zap.String("body", event.GetIssue().GetBody()))
		err := pr.HandleIssueCommentEvent(ctx, client, event)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("error handling issue comment", zap.Error(err), zap.Any("event", event))
		}
	}
}

func (r *Registry) CallCommitCommentPlugins(ctx context.Context, client *github.Client, event *github.CommitCommentEvent) {
	for _, pr := range r.ccplugins {
		contextutils.LoggerFrom(ctx).Debugw("Issue comment",
			zap.String("owner", event.GetRepo().GetOwner().GetLogin()),
			zap.String("repo", event.GetRepo().GetName()),
			zap.String("body", event.GetComment().GetBody()))
		err := pr.HandleCommitCommentEvent(ctx, client, event)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("error handling commit comment", zap.Error(err), zap.Any("event", event))
		}
	}
}

func (r *Registry) CallReleasePlugins(ctx context.Context, client *github.Client, event *github.ReleaseEvent) {
	for _, pr := range r.releaseplugins {
		contextutils.LoggerFrom(ctx).Debugw("Release",
			zap.String("tag", event.GetRelease().GetTagName()),
			zap.String("org", event.GetRepo().GetOwner().GetLogin()),
			zap.String("repo", event.GetRepo().GetName()))
		err := pr.HandleReleaseEvent(ctx, client, event)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("error handling release", zap.Error(err), zap.Any("event", event))
		}
	}
}

func (r *Registry) CallIssuesPlugins(ctx context.Context, client *github.Client, event *github.IssuesEvent) {
	for _, pr := range r.issuesplugins {
		err := pr.HandleIssuesEvent(ctx, client, event)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("error handling issues", zap.Error(err), zap.Any("event", event))
		}
	}
}
