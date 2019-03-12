package githubutils

import (
	"context"
	"github.com/google/go-github/github"
	"github.com/solo-io/go-utils/logger"
)

func CommentAlreadyExists(ctx context.Context, client *github.Client, owner, repo string, prNumber int, commentBody string) bool {
	opts := github.IssueListCommentsOptions{}
	comments, _, err := client.Issues.ListComments(ctx, owner, repo, prNumber, &opts)
	if err != nil {
		logger.Warnf("Could not list comments, error was: %v", err)
		return false
	}
	for _, comment := range comments {
		if comment.GetBody() == commentBody {
			return true
		}
	}
	return false
}
