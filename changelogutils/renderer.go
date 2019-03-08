package changelogutils

type ChangelogRenderer interface {
	Render(changelog *Changelog) string
}

type MarkdownChangelogRenderer struct{}

func (renderer *MarkdownChangelogRenderer) Render(changelog *Changelog) string {
	return GenerateChangelogMarkdown(changelog)
}
