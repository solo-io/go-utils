package issuewriter

import (
	"context"

	"github.com/google/go-github/v32/github"
)

// IssueWriter writes the generated contents of a scan to a location, either a file on the local filesystem
// or a GitHub issue.
type IssueWriter interface {
	// Write writes `contents`, the results of a scan of the images in `release`, to a location
	// designated by the implementation.
	Write(ctx context.Context, release *github.RepositoryRelease, contents string) error
}
