package issuewriter

import (
	"context"

	"github.com/google/go-github/v32/github"
)

type IssueWriter interface {
	Write(ctx context.Context, release *github.RepositoryRelease, contents string) error
}
