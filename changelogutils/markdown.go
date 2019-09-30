package changelogutils

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/afero"

	"github.com/solo-io/go-utils/vfsutils"
)

/*
Changelog markdown:

summary
breaking changes
new features
fixes
closing

*/

func GenerateChangelogFromLocalDirectory(ctx context.Context, repoRootPath, owner, repo, sha, changelogDirPath string, w io.Writer) error {
	fs := afero.NewOsFs()
	mountedRepo := vfsutils.NewLocalMountedRepoForFs(fs, repoRootPath, owner, repo, sha)
	dirContent, err := fs.Open(changelogDirPath)
	if err != nil {
		return err
	}
	fmt.Println(dirContent)
	dirs, err := dirContent.Readdirnames(-1)
	fmt.Println(dirs)
	if err != nil {
		return err
	}
	reader := NewChangelogReader(mountedRepo)
	return GenerateChangelogForTags(ctx, dirs, reader, w)

}
func GenerateChangelogForTags(ctx context.Context, tags []string, reader ChangelogReader, w io.Writer) error {
	changelogs := make(ChangelogList, len(tags))
	var err error
	for i, tag := range tags {
		if changelogs[i], err = reader.GetChangelogForTag(ctx, tag); err != nil {
			return err
		}
	}
	sort.Sort(changelogs)
	for _, cl := range changelogs {
		md := GenerateChangelogMarkdown(cl)
		if _, err := fmt.Fprintf(w, "# %v\n", cl.Version.String()); err != nil {
			return err
		}
		if _, err := fmt.Fprint(w, md); err != nil {
			return err
		}
	}
	return nil
}
func GenerateChangelogMarkdown(changelog *Changelog) string {
	output := changelog.Summary
	if output != "" {
		output = output + "\n\n"
	}

	dependencyBumps := renderDependencyBumps(changelog)
	if dependencyBumps != "" {
		output = output + "**Dependency Bumps**\n\n" + dependencyBumps + "\n"
	}

	breakingChanges := renderChangelogEntries(changelog, BREAKING_CHANGE)
	if breakingChanges != "" {
		output = output + "**Breaking Changes**\n\n" + breakingChanges + "\n"
	}

	newFeatures := renderChangelogEntries(changelog, NEW_FEATURE)
	if newFeatures != "" {
		output = output + "**New Features**\n\n" + newFeatures + "\n"
	}

	fixes := renderChangelogEntries(changelog, FIX)
	if fixes != "" {
		output = output + "**Fixes**\n\n" + fixes + "\n"
	}

	if changelog.Closing != "" {
		output = output + changelog.Closing + "\n\n"
	}

	if output == "" {
		output = "This release contained no user-facing changes.\n\n"
	}
	return output
}

func renderDependencyBumps(changelog *Changelog) string {
	output := ""
	for _, file := range changelog.Files {
		for _, entry := range file.Entries {
			if entry.Type == DEPENDENCY_BUMP {
				output = output + "- " + entry.DependencyOwner + "/" + entry.DependencyRepo + " has been upgraded to " + entry.DependencyTag + ".\n"
			}
		}
	}
	return output
}

func renderChangelogEntries(changelog *Changelog, entryType ChangelogEntryType) string {
	output := ""
	for _, file := range changelog.Files {
		for _, entry := range file.Entries {
			if entry.Type == entryType {
				output = output + renderChangelogEntry(entry) + "\n"
			}
		}
	}
	return output
}

func renderChangelogEntry(entry *ChangelogEntry) string {
	description := strings.TrimSpace(entry.Description)
	link := strings.TrimSpace(entry.IssueLink)
	return "- " + description + " (" + link + ")"
}
