package changelogutils

import (
	"context"
	"golang.org/x/mod/semver"
	"html/template"
	"io"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/solo-io/go-utils/vfsutils"
)

/*
Changelog markdown:

summary
dependency bumps
breaking changes
upgrade
helm
new features
fixes
closing

*/

var (
	MountLocalDirectoryError = func(err error) error {
		return errors.Wrapf(err, "unable to mount local directory")
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
	mountedRepo, err := vfsutils.NewLocalMountedRepoForFs(repoRootPath, owner, repo)
	if err != nil {
		return MountLocalDirectoryError(err)
	}
	files, err := mountedRepo.ListFiles(ctx, changelogDirPath)
	if err != nil {
		return ReadChangelogDirError(err)
	}
	var tags []string
	for _, file := range files {
		if file.IsDir() {
			tags = append(tags, file.Name())
		}
	}
	reader := NewChangelogReader(mountedRepo)
	return GenerateChangelogForTags(ctx, tags, reader, w)
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

	upgradeNotes := renderChangelogEntries(changelog, UPGRADE)
	if upgradeNotes != "" {
		output = output + "**Upgrade Notes**\n\n" + upgradeNotes + "\n"
	}

	helmChanges := renderChangelogEntries(changelog, HELM)
	if helmChanges != "" {
		output = output + "**Helm Changes**\n\n" + helmChanges + "\n"
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
	maxDependencyMap := make(map[string]string)
	// Using a set instead of a slice to avoid duplicate entries on non-semantic version bumps
	type set map[string]struct{}
	nonSemanticVersionMap := make(map[string]set)
	for _, file := range changelog.Files {
		for _, entry := range file.Entries {
			if entry.Type == DEPENDENCY_BUMP {
				dependency := entry.DependencyOwner + "/" + entry.DependencyRepo
				// If the tag is not a valid semantic version, we can't compare it to other tags, so we will output all
				if !semver.IsValid(entry.DependencyTag) {
					if _, ok := nonSemanticVersionMap[dependency]; ok {
						nonSemanticVersionMap[dependency][entry.DependencyTag] = struct{}{}
					} else {
						nonSemanticVersionMap[dependency] = set{entry.DependencyTag: struct{}{}}
					}
				} else {
					if val, ok := maxDependencyMap[dependency]; ok {
						// Note: A limitation with semver.Compare is that postfixes are compared as strings.
						// For example, an issue is the comparison between "v1.0.0-patch1" and "v1.0.0-rc1".
						// This would result that "v1.0.0-patch1" is greater than "v1.0.0-rc1", which is technically not true.
						// TODO: We could set a priority (ex. `rc` > `patch` > `beta`) and do a second comparison through that.
						if semver.Compare(entry.DependencyTag, val) > 0 {
							maxDependencyMap[dependency] = entry.DependencyTag
						}
					} else {
						maxDependencyMap[dependency] = entry.DependencyTag
					}
				}
			}
		}
	}

	var semanticKeys []string
	for k := range maxDependencyMap {
		semanticKeys = append(semanticKeys, k)
	}
	var nonSemanticKeys []string
	for k := range nonSemanticVersionMap {
		nonSemanticKeys = append(nonSemanticKeys, k)
	}
	sort.Strings(semanticKeys)
	sort.Strings(nonSemanticKeys)

	output := ""
	var semanticKeyIndex, nonSemanticKeyIndex int
	for semanticKeyIndex < len(semanticKeys) && nonSemanticKeyIndex < len(nonSemanticKeys) {
		if semanticKeys[semanticKeyIndex] < nonSemanticKeys[nonSemanticKeyIndex] {
			output = output + "- " + semanticKeys[semanticKeyIndex] + " has been upgraded to " + maxDependencyMap[semanticKeys[semanticKeyIndex]] + ".\n"
			semanticKeyIndex++
		} else {
			for dependencyTag, _ := range nonSemanticVersionMap[nonSemanticKeys[nonSemanticKeyIndex]] {
				output = output + "- " + nonSemanticKeys[nonSemanticKeyIndex] + " has been upgraded to " + dependencyTag + ".\n"
			}
			nonSemanticKeyIndex++
		}
	}
	for _, key := range semanticKeys[semanticKeyIndex:] {
		output = output + "- " + key + " has been upgraded to " + maxDependencyMap[key] + ".\n"
	}
	for _, key := range nonSemanticKeys[nonSemanticKeyIndex:] {
		for dependencyTag, _ := range nonSemanticVersionMap[key] {
			output = output + "- " + key + " has been upgraded to " + dependencyTag + ".\n"
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
### {{ .ReleaseVersionString }}

{{ .Summary }}
{{- end -}}
`))
