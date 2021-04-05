package changelogdocutils

import (
	"fmt"
	"github.com/google/go-github/v32/github"
)

/**
This package is meant to provide utilities for generating organized changelog docs for the docs site.
Outputs changelog doc markdown.
*/

type Options struct {
}

/**
Changelog formats
*/
const (
	Chronological = iota
	// Generates changelogs in order of minor release (v1.8 - v1.8.0-beta, v1.8.1)
	GroupedByMinorRelease
	// Groups all changelogs per minor release (v1.8.0, v1.7.0...)
	MinorRelease
)

func GetGithubReleaseMarkdownLink(release *github.RepositoryRelease, repoOwner, repo string) string {
	tag := release.GetTagName()
	link := fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", repoOwner, repo, tag)
	return Link(tag, link)
}

type ChangelogGenerator interface {
	Generate() (string, error)
}