package botutils

import (
	"context"
	"github.com/solo-io/go-utils/botutils/botconfig"

	"github.com/google/go-github/github"
)

type Plugin interface {
}

type PullRequestHandler interface {
	Plugin
	HandlePREvent(ctx context.Context, client *github.Client, config *botconfig.ApplicationConfig, event *github.PullRequestEvent) error
}

type PullRequestReviewHandler interface {
	Plugin
	HandlePullRequestReviewEvent(ctx context.Context, client *github.Client, config *botconfig.ApplicationConfig, event *github.PullRequestReviewEvent) error
}

type IssueCommentHandler interface {
	Plugin
	HandleIssueCommentEvent(ctx context.Context, client *github.Client, config *botconfig.ApplicationConfig, event *github.IssueCommentEvent) error
}

type CommitCommentHandler interface {
	Plugin
	HandleCommitCommentEvent(ctx context.Context, client *github.Client, config *botconfig.ApplicationConfig, event *github.CommitCommentEvent) error
}

type ReleaseHandler interface {
	Plugin
	HandleReleaseEvent(ctx context.Context, client *github.Client, config *botconfig.ApplicationConfig, event *github.ReleaseEvent) error
}
