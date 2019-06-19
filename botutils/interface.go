package botutils

import (
	"context"

	"github.com/google/go-github/github"
)

type Plugin interface {
}

type PullRequestHandler interface {
	Plugin
	HandlePREvent(ctx context.Context, client *github.Client, fetchedConfig *FetchedConfig, event *github.PullRequestEvent) error
}

type PullRequestReviewHandler interface {
	Plugin
	HandlePullRequestReviewEvent(ctx context.Context, client *github.Client, fetchedConfig *FetchedConfig, event *github.PullRequestReviewEvent) error
}

type IssueCommentHandler interface {
	Plugin
	HandleIssueCommentEvent(ctx context.Context, client *github.Client, fetchedConfig *FetchedConfig, event *github.IssueCommentEvent) error
}

type CommitCommentHandler interface {
	Plugin
	HandleCommitCommentEvent(ctx context.Context, client *github.Client, fetchedConfig *FetchedConfig, event *github.CommitCommentEvent) error
}

type ReleaseHandler interface {
	Plugin
	HandleReleaseEvent(ctx context.Context, client *github.Client, fetchedConfig *FetchedConfig, event *github.ReleaseEvent) error
}
