package changelogutils

import (
	"context"
	"html/template"
	"io"
	"sort"
	"strings"

	"github.com/pkg/errors"

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

var (
	MountLocalDirectoryError = func(err error) error {
		return errors.Wrapf(err, "unable to mount local directory")
	}
	OpenChangelogDirError = func(err error) error {
		return errors.Wrapf(err, "unable to open changelog directory")
	}
	ReadChangelogDirError = func(err error) error {
		return errors.Wrapf(err, "unable to read changelog directory")
	}
	GetChangelogForTagError = func(err error) error {
		return errors.Wrapf(err, "unable to get changelog for tag")
	}
	GenerateChangelogSummaryTemplateError = func(err error) error {
		return errors.Wrapf(err, "unable to generate changelog summary from template")
	}
)

func GenerateChangelogFromLocalDirectory(ctx context.Context, repoRootPath, owner, repo, changelogDirPath string, w io.Writer) error {
	fs := afero.NewOsFs()
	mountedRepo, err := vfsutils.NewLocalMountedRepoForFs(fs, repoRootPath, owner, repo)
	if err != nil {
		return MountLocalDirectoryError(err)
	}
	dirContent, err := fs.Open(changelogDirPath)
	if err != nil {
		return OpenChangelogDirError(err)
	}
	dirs, err := dirContent.Readdirnames(-1)
	if err != nil {
		return ReadChangelogDirError(err)
	}
	reader := NewChangelogReader(mountedRepo)
	return GenerateChangelogForTags(ctx, dirs, reader, w)
}

func GenerateChangelogForTags(ctx context.Context, tags []string, reader ChangelogReader, w io.Writer) error {
	changelogs := make(ChangelogList, len(tags))
	var err error
	for i, tag := range tags {
		if changelogs[i], err = reader.GetChangelogForTag(ctx, tag); err != nil {
			return GetChangelogForTagError(err)
		}
	}
	sort.Sort(sort.Reverse(changelogs))
	tmplData := changelogSummaryTmplDataFromChangelogs(changelogs)
	if err := changelogSummaryTmpl.Execute(w, tmplData); err != nil {
		return GenerateChangelogSummaryTemplateError(err)
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

type ChangelogTmplData struct {
	ReleaseVersionString string
	Summary              string
}

func changelogSummaryTmplDataFromChangelogs(changelogs ChangelogList) []ChangelogTmplData {
	d := make([]ChangelogTmplData, len(changelogs))
	for i, c := range changelogs {
		md := GenerateChangelogMarkdown(c)
		d[i] = ChangelogTmplData{
			ReleaseVersionString: c.Version.String(),
			Summary:              md,
		}
	}
	return d
}

var changelogSummaryTmpl = template.Must(
	template.New("changelog summary").Parse(`
{{ range . }}
# {{ .ReleaseVersionString }}

{{ .Summary }}
{{- end -}}
`))
