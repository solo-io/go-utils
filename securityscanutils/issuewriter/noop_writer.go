package issuewriter

import (
	"context"

	"github.com/google/go-github/v32/github"
)

// NoopWriter provides a no-op implementation of the IssueWriter interface, used when the
// specified scan action is `none`.
type NoopWriter struct{}

var _ IssueWriter = &NoopWriter{}

func NewNoopWriter() IssueWriter {
	return &NoopWriter{}
}

func (n *NoopWriter) Write(_ context.Context, _ *github.RepositoryRelease, _ string) error {
	return nil
}
