package changelogutils

import (
	"context"
	"github.com/ghodss/yaml"
	"github.com/solo-io/go-utils/githubutils"
)

type ChangelogEntryType int

const (
	BREAKING_CHANGE ChangelogEntryType = iota
	FIX
	NEW_FEATURE
	NON_USER_FACING
)

type ChangelogEntry struct {
	Type        ChangelogEntryType
	Description string
}

type ChangelogFile struct {
	Entries []ChangelogEntry `json:"changelog,omitempty"`
}

type Changelog struct {
	Files []ChangelogFile
	Summary string
	Version string
}

type RawChangelogFile struct {
	Filename string
	Bytes []byte
}

const (
	ChangelogDirectory = "changelog"
	Master = "master"
)

// Should return the last released version
// Executes git commands, so this won't work if running from an unzipped archive of the code.
func GetLatestTag(ctx context.Context, owner, repo string) (string, error) {
	client, err := githubutils.GetClient(ctx)
	if err != nil {
		return "", err
	}
	return githubutils.FindLatestReleaseTag(ctx, client, owner, repo)
}



func BytesToChangelog(rcl []*RawChangelogFile) (*Changelog, error) {
	cl := &Changelog{
		Files: make([]ChangelogFile, len(rcl)),
	}
	for _, file := range rcl {
		var clf ChangelogFile
		err := yaml.Unmarshal(file.Bytes, &clf)
		if err != nil {
			return nil, err
		}
		cl.Files = append(cl.Files, clf)
	}
	return cl, nil
}
