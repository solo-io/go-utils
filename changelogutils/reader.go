package changelogutils

import (
	"context"
	"github.com/google/go-github/github"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/versionutils"
	"path/filepath"
)

type ChangelogReader interface {
	ReadChangelogFile(owner, repo, ref, path string) (*ChangelogFile, error)
	ReadChangelogForTag(owner, repo, ref, tag string) (*Changelog, error)
	GetProposedChangelog(owner, repo, ref string) (*Changelog, error)
}

type githubChangelogReader struct {
	ctx    context.Context
	client *github.Client
	parser ChangelogParser
}

func NewGithubChangelogReader(ctx context.Context) (*githubChangelogReader, error) {
	client, err := githubutils.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	return &githubChangelogReader{
		ctx:    ctx,
		client: client,
		parser: NewChangelogParser(),
	}, nil
}

func (reader *githubChangelogReader) ReadChangelogFile(owner, repo, ref, path string) (*ChangelogFile, error) {
	contents, err := reader.readFile(owner, repo, ref, path)
	if err != nil {
		return nil, err
	}
	return reader.parser.ParseChangelogFile(contents)
}

func (reader *githubChangelogReader) ReadChangelogForTag(owner, repo, ref, tag string) (*Changelog, error) {
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

func (reader *githubChangelogReader) readFile(owner, repo, ref, path string) (string, error) {
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

func (reader *githubChangelogReader) GetProposedChangelog(owner, repo, ref string) (*Changelog, error) {
	proposedTag, err := reader.getProposedTag(owner, repo, ref)
	if err != nil {
		return nil, err
	}
	return reader.ReadChangelogForTag(owner, repo, ref, proposedTag)
}

func (reader *githubChangelogReader) getProposedTag(owner, repo, ref string) (string, error) {
	opts := &github.RepositoryContentGetOptions{
		Ref: ref,
	}
	_, directoryContent, _, err := reader.client.Repositories.GetContents(reader.ctx, owner, repo, ChangelogDirectory, opts)
	if err != nil {
		return "", err
	}
	proposedTag := ""
	latestTag, err := githubutils.FindLatestReleaseTagIncudingPrerelease(reader.ctx, reader.client, owner, repo)
	if err != nil {
		return "", err
	}
	for _, subdirectory := range directoryContent {
		if subdirectory.GetType() != "dir" {
			return "", errors.Errorf("Expected contents of changelog to be type dir, found %s of type %s", subdirectory.GetName(), subdirectory.GetType())
		}
		if !versionutils.MatchesRegex(subdirectory.GetName()) {
			return "", newErrorInvalidDirectoryName(subdirectory.GetName())
		}
		greaterThan, err := versionutils.IsGreaterThanTag(subdirectory.GetName(), latestTag)
		if err != nil {
			return "", err
		}
		if greaterThan {
			if proposedTag != "" {
				return "", newErrorMultipleVersionsFound(subdirectory.GetName(), proposedTag, latestTag)
			}
			proposedTag = subdirectory.GetName()
		}
	}
	if proposedTag == "" {
		return "", newErrorNoVersionFound(latestTag)
	}
	return proposedTag, nil
}