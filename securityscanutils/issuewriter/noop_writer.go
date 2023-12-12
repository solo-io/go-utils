package issuewriter

import (
	"context"
	"github.com/google/go-github/v32/github"
)

type NoopWriter struct{}

var nw IssueWriter = &NoopWriter{}

func NewNoopWriter() IssueWriter {
	return &NoopWriter{}
}

func (n *NoopWriter) Write(ctx context.Context, release *github.RepositoryRelease, contents string) error {
	return nil
}
