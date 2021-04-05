package changelogdocutils

import "fmt"

func H1 (s string) string{
	return fmt.Sprintf("\n# %s\n", s)
}

func H2 (s string) string{
	return fmt.Sprintf("\n## %s\n", s)
}

func H3 (s string) string{
	return fmt.Sprintf("\n### %s\n", s)
}

func H4 (s string) string{
	return fmt.Sprintf("\n#### %s\n", s)
}

func H5 (s string) string{
	return fmt.Sprintf("\n##### %s\n", s)
}

func Bold (s string) string{
	return fmt.Sprintf("**%s**", s)
}

func Italic (s string) string{
	return fmt.Sprintf("*%s*", s)
}

func OrderedListItem (s string) string {
	return fmt.Sprintf("1. %s\n", s)
}

func UnorderedListItem (s string) string {
	return fmt.Sprintf("- %s\n", s)
}

func Link(title, link string) string {
	return fmt.Sprintf("[%s](%s)", title, link)
}