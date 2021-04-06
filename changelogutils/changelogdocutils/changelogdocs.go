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


func GetGithubReleaseMarkdownLink(release *github.RepositoryRelease, repoOwner, repo string) string {
	tag := release.GetTagName()
	link := fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", repoOwner, repo, tag)
	return Link(tag, link)
}

type ChangelogGenerator interface {
	Generate() (string, error)
	GenerateJSON() (string, error)
}