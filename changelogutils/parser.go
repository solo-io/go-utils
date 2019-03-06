package changelogutils

import (
	"github.com/ghodss/yaml"
	"github.com/solo-io/go-utils/errors"
)

type ChangelogParser interface {
	ParseChangelogFile(contents string) (*ChangelogFile, error)
}

type changelogParser struct {}

func NewChangelogParser() ChangelogParser {
	return &changelogParser{}
}

func (parser *changelogParser) ParseChangelogFile(contents string) (*ChangelogFile, error) {
	var changelogFile ChangelogFile
	if err := yaml.Unmarshal([]byte(contents), &changelogFile); err != nil {
		return nil, errors.Errorf("String could not be parsed as a changelog file. Error: %v", err)
	}

	for _, entry := range changelogFile.Entries {
		if entry.Type != NON_USER_FACING && entry.Type != DEPENDENCY_BUMP {
			if entry.IssueLink == "" {
				return nil, errors.Errorf("Changelog entries must have an issue link")
			}
			if entry.Description == "" {
				return nil, errors.Errorf("Changelog entries must have a description")
			}
		}
		if entry.Type == DEPENDENCY_BUMP {
			if entry.DependencyOwner == "" {
				return nil, errors.Errorf("Dependency bumps must have an owner")
			}
			if entry.DependencyRepo == "" {
				return nil, errors.Errorf("Dependency bumps must have a repo")
			}
			if entry.DependencyTag == "" {
				return nil, errors.Errorf("Dependency bumps must have a tag")
			}
		}
	}

	return &changelogFile, nil
}
