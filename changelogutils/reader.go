package changelogutils

import (
	"context"
	"github.com/google/go-github/github"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/versionutils"
	"path/filepath"
	"sort"
)

type ChangelogReader interface {
	GetAllChangelogVersionsDesc(owner, repo, ref string) ([]*versionutils.Version, error)
	ReadChangelogFile(owner, repo, ref, path string) (*ChangelogFile, error)
	ReadChangelogForTag(owner, repo, ref, tag string) (*Changelog, error)
	GetProposedChangelog(owner, repo, ref string) (*versionutils.Version, *Changelog, error)
}

type githubChangelogReader struct {
	ctx    context.Context
	client *github.Client
	parser ChangelogParser
}

func NewChangelogReader(ctx context.Context) (ChangelogReader, error) {
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

func (reader *githubChangelogReader) GetProposedChangelog(owner, repo, ref string) (*versionutils.Version, *Changelog, error) {
	proposedVersion, err := reader.getProposedVersion(owner, repo, ref)
	if err != nil {
		return nil, nil, err
	}
	changelog, err := reader.ReadChangelogForTag(owner, repo, ref, proposedVersion.String())
	if err != nil {
		return nil, nil, err
	}
	return proposedVersion, changelog, nil
}

func (reader *githubChangelogReader) getProposedVersion(owner, repo, ref string) (*versionutils.Version, error) {
	opts := &github.RepositoryContentGetOptions{
		Ref: ref,
	}
	_, directoryContent, _, err := reader.client.Repositories.GetContents(reader.ctx, owner, repo, ChangelogDirectory, opts)
	if err != nil {
		return nil, err
	}
	proposedTag := ""
	latestTag, err := githubutils.FindLatestReleaseTagIncudingPrerelease(reader.ctx, reader.client, owner, repo)
	if err != nil {
		return nil, err
	}
	for _, subdirectory := range directoryContent {
		if subdirectory.GetType() != "dir" {
			return nil, errors.Errorf("Expected contents of changelog to be type dir, found %s of type %s", subdirectory.GetName(), subdirectory.GetType())
		}
		if !versionutils.MatchesRegex(subdirectory.GetName()) {
			return nil, newErrorInvalidDirectoryName(subdirectory.GetName())
		}
		greaterThan, err := versionutils.IsGreaterThanTag(subdirectory.GetName(), latestTag)
		if err != nil {
			return nil, err
		}
		if greaterThan {
			if proposedTag != "" {
				return nil, newErrorMultipleVersionsFound(subdirectory.GetName(), proposedTag, latestTag)
			}
			proposedTag = subdirectory.GetName()
		}
	}
	if proposedTag == "" {
		return nil, newErrorNoVersionFound(latestTag)
	}
	proposedVersion, err := versionutils.ParseVersion(proposedTag)
	if err != nil {
		return nil, err
	}
	return proposedVersion, nil
}

func (reader *githubChangelogReader) GetAllChangelogVersionsDesc(owner, repo, ref string) ([]*versionutils.Version, error) {
	opts := github.RepositoryContentGetOptions{
		Ref: ref,
	}
	_, contents, _, err := reader.client.Repositories.GetContents(reader.ctx, owner, repo, ChangelogDirectory, &opts)
	if err != nil {
		return nil, err
	}
	var versions []*versionutils.Version
	for _, subdirectory := range contents {
		if subdirectory.GetType() != "dir" {
			return nil, errors.Errorf("Expected contents of changelog to be type dir, found %s of type %s", subdirectory.GetName(), subdirectory.GetType())
		}
		version, err := versionutils.ParseVersion(subdirectory.GetName())
		if err != nil {
			return nil, newErrorInvalidDirectoryName(subdirectory.GetName())
		}
		versions = append(versions, version)
	}
	sort.Sort(versionutils.ByVersionDesc(versions))
	return versions, nil
}
