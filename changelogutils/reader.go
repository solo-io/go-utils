package changelogutils

import (
	"context"
	"github.com/google/go-github/github"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/versionutils"
	"path/filepath"
)

type ChangelogReader interface {
	ReadChangelogFile(owner, repo, ref, path string) (*ChangelogFile, error)
	ReadChangelogForTag(owner, repo, ref, tag string) (*Changelog, error)
}

type GithubChangelogReader struct {
	ctx    context.Context
	client *github.Client
	parser ChangelogParser
}

func NewGithubChangelogReader(ctx context.Context) (*GithubChangelogReader, error) {
	client, err := githubutils.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	return &GithubChangelogReader{
		ctx:    ctx,
		client: client,
		parser: NewChangelogParser(),
	}, nil
}

func (reader *GithubChangelogReader) ReadChangelogFile(owner, repo, ref, path string) (*ChangelogFile, error) {
	contents, err := reader.readFile(owner, repo, ref, path)
	if err != nil {
		return nil, err
	}
	return reader.parser.ParseChangelogFile(contents)
}

func (reader *GithubChangelogReader) ReadChangelog(owner, repo, ref, tag string) (*Changelog, error) {
	version, err := versionutils.ParseVersion(tag)
	if err != nil {
		return nil, err
	}
	opts := &github.RepositoryContentGetOptions{
		Ref: ref,
	}
	directory := filepath.Join(ChangelogDirectory, tag)
	_, directoryContent, _, err := reader.client.Repositories.GetContents(reader.ctx, owner, repo, directory, opts)
	if err != nil {
		return nil, err
	}
	changelog := Changelog{
		Version: version,
	}
	for _, contentFile := range directoryContent {
		contents, err := reader.readFile(owner, repo, ref, contentFile.GetPath())
		if err != nil {
			return nil, err
		}
		if contentFile.GetName() == SummaryFile {
			changelog.Summary = contents
		} else if contentFile.GetName() == ClosingFile {
			changelog.Closing = contents
		} else {
			changelogFile, err := reader.parser.ParseChangelogFile(contents)
			if err != nil {
				return nil, err
			}
			changelog.Files = append(changelog.Files, changelogFile)
		}
	}
	return &changelog, nil
}

func (reader *GithubChangelogReader) readFile(owner, repo, ref, path string) (string, error) {
	opts := &github.RepositoryContentGetOptions{
		Ref: ref,
	}
	fileContent, _, _, err := reader.client.Repositories.GetContents(reader.ctx, owner, repo, path, opts)
	content, err := fileContent.GetContent()
	if err != nil {
		return "", err
	}
	return content, nil
}
