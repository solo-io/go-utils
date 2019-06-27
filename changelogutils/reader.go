package changelogutils

import (
	"context"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/versionutils"
	"github.com/solo-io/go-utils/vfsutils"
	"path/filepath"
)

//go:generate mockgen -destination reader_mock_test.go -self_package github.com/solo-io/go-utils/changelogutils -package changelogutils_test github.com/solo-io/go-utils/vfsutils MountedRepo

var (
	UnableToListFilesError = func(err error, directory string) error {
		return errors.Wrapf(err, "Unable to list files in directory %s", directory)
	}
	UnexpectedDirectoryError = func(name, directory string) error {
		return errors.Errorf("Unexpected directory %s in changelog directory %s", name, directory)
	}
	UnableToReadSummaryFileError = func(err error, path string) error {
		return errors.Wrapf(err, "Unable to read summary file %s", path)
	}
	UnableToReadClosingFileError = func(err error, path string) error {
		return errors.Wrapf(err, "Unable to read closing file %s", path)
	}
	NoEntriesInChangelogError = func(filename string) error {
		return errors.Errorf("No changelog entries found in file %s.", filename)
	}
	UnableToParseChangelogError = func(err error, path string) error {
		return errors.Wrapf(err, "File %s is not a valid changelog file.", path)
	}
	MissingIssueLinkError = errors.Errorf("Changelog entries must have an issue link")
	MissingDescriptionError = errors.Errorf("Changelog entries must have a description")
	MissingOwnerError = errors.Errorf("Dependency bumps must have an owner")
	MissingRepoError = errors.Errorf("Dependency bumps must have a repo")
	MissingTagError = errors.Errorf("Dependency bumps must have a tag")
)

type ChangelogReader interface {
	GetChangelogForTag(ctx context.Context, tag string) (*Changelog, error)
	ReadChangelogFile(ctx context.Context, path string) (*ChangelogFile, error)
}

type changelogReader struct {
	code vfsutils.MountedRepo
}

func NewChangelogReader(code vfsutils.MountedRepo) ChangelogReader {
	return &changelogReader{code: code}
}

func (c* changelogReader) GetChangelogForTag(ctx context.Context, tag string) (*Changelog, error) {
	version, err := versionutils.ParseVersion(tag)
	if err != nil {
		return nil, err
	}
	changelog := Changelog{
		Version: version,
	}
	changelogPath := filepath.Join(ChangelogDirectory, tag)
	files, err := c.code.ListFiles(ctx, changelogPath)
	if err != nil {
		return nil, UnableToListFilesError(err, changelogPath)
	}
	for _, changelogFileInfo := range files {
		if changelogFileInfo.IsDir() {
			return nil, UnexpectedDirectoryError(changelogFileInfo.Name(), changelogPath)
		}
		changelogFilePath := filepath.Join(changelogPath, changelogFileInfo.Name())
		if changelogFileInfo.Name() == SummaryFile {
			summary, err := c.code.GetFileContents(ctx, changelogFilePath)
			if err != nil {
				return nil, UnableToReadSummaryFileError(err, changelogFilePath)
			}
			changelog.Summary = string(summary)
		} else if changelogFileInfo.Name() == ClosingFile {
			closing, err := c.code.GetFileContents(ctx, changelogFilePath)
			if err != nil {
				return nil, UnableToReadClosingFileError(err, changelogFilePath)
			}
			changelog.Closing = string(closing)
		} else {
			changelogFile, err := c.ReadChangelogFile(ctx, changelogFilePath)
			if err != nil {
				return nil, err
			}
			changelog.Files = append(changelog.Files, changelogFile)
		}
	}

	return &changelog, nil
}

func (c* changelogReader) ReadChangelogFile(ctx context.Context, path string) (*ChangelogFile, error) {
	var changelog ChangelogFile
	bytes, err := c.code.GetFileContents(ctx, path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(bytes, &changelog); err != nil {
		return nil, UnableToParseChangelogError(err, path)
	}

	if len(changelog.Entries) == 0 {
		return nil, NoEntriesInChangelogError(path)
	}

	for _, entry := range changelog.Entries {
		if entry.Type != NON_USER_FACING && entry.Type != DEPENDENCY_BUMP {
			if entry.IssueLink == "" {
				return nil, MissingIssueLinkError
			}
			if entry.Description == "" {
				return nil, MissingDescriptionError
			}
		}
		if entry.Type == DEPENDENCY_BUMP {
			if entry.DependencyOwner == "" {
				return nil, MissingOwnerError
			}
			if entry.DependencyRepo == "" {
				return nil, MissingRepoError
			}
			if entry.DependencyTag == "" {
				return nil, MissingTagError
			}
		}
	}

	return &changelog, nil
}
