package issuewriter

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v32/github"
)

// LocalIssueWriter writes the scan results to a file on the local file system.
type LocalIssueWriter struct {
	// The details about the GitHub repository
	repo GithubRepo

	// The directory in which to create files
	outputDir string
}

var _ IssueWriter = &LocalIssueWriter{}

func NewLocalIssueWriter(repo GithubRepo, outputDir string) (IssueWriter, error) {
	// Set up the directory structure for local output
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return &LocalIssueWriter{
		repo:      repo,
		outputDir: outputDir,
	}, nil
}

func (l *LocalIssueWriter) Write(_ context.Context, release *github.RepositoryRelease, contents string) error {
	version, err := semver.NewVersion(release.GetTagName())
	if err != nil {
		return err
	}
	filename := path.Join(l.outputDir, version.String()+".md")
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, contents)
	if err != nil {
		return err
	}
	return nil
}
