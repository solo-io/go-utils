package changelogutils

import "strings"

type ChangelogRenderer interface {
	Render(changelog *Changelog) string
}

func GenerateChangelogMarkdown(changelog *Changelog) string {
	renderer := markdownChangelogRenderer{}
	return renderer.Render(changelog)
}

func NewDefaultChangelogRenderer() ChangelogRenderer {
	return &markdownChangelogRenderer{}
}

type markdownChangelogRenderer struct{}

func (renderer *markdownChangelogRenderer) Render(changelog *Changelog) string {
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