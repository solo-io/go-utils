package botutils

import (
	"context"

	"github.com/google/go-github/v32/github"
)

type Plugin interface {
}

type PullRequestHandler interface {
	Plugin
	HandlePREvent(ctx context.Context, client *github.Client, event *github.PullRequestEvent) error
}

type PullRequestReviewHandler interface {
	Plugin
	HandlePullRequestReviewEvent(ctx context.Context, client *github.Client, event *github.PullRequestReviewEvent) error
}

type IssueCommentHandler interface {
	Plugin
	HandleIssueCommentEvent(ctx context.Context, client *github.Client, event *github.IssueCommentEvent) error
}

type CommitCommentHandler interface {
	Plugin
	HandleCommitCommentEvent(ctx context.Context, client *github.Client, event *github.CommitCommentEvent) error
}

type ReleaseHandler interface {
	Plugin
	HandleReleaseEvent(ctx context.Context, client *github.Client, event *github.ReleaseEvent) error
}

type IssuesHandler interface {
	Plugin
	HandleIssuesEvent(ctx context.Context, client *github.Client, event *github.IssuesEvent) error
}
