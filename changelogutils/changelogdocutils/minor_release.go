package changelogdocutils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v32/github"
	errors "github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/githubutils"
	. "github.com/solo-io/go-utils/versionutils"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"sort"
	"strings"
)

func UnableToGenerateChangelogError(err error) error {
	return errors.Wrap(err, "Unable to generate changelog")
}
func UnableToParseVersionError(err error, versionTag string) error {
	return errors.Wrapf(err, "Unable to parse version tag %s", versionTag)
}

type MinorReleaseGroupedChangelogGenerator struct {
	Client    *github.Client
	RepoOwner string
	Repo      string
}

func NewMinorReleaseGroupedChangelogGenerator(client *github.Client, repoOwner, repo string) *MinorReleaseGroupedChangelogGenerator {
	return &MinorReleaseGroupedChangelogGenerator{
		Client:    client,
		RepoOwner: repoOwner,
		Repo:      repo,
	}
}

func (g *MinorReleaseGroupedChangelogGenerator) Generate(ctx context.Context) (string, error) {
	releaseData, err := g.GetReleaseData(ctx)
	if err != nil {
		return "", err
	}
	return releaseData.String(), nil
}

func (g *MinorReleaseGroupedChangelogGenerator) GetReleaseData(ctx context.Context) (*ReleaseData, error) {
	releases, err := githubutils.GetAllRepoReleases(ctx, g.Client, g.RepoOwner, g.Repo)
	if err != nil {
		return nil, err
	}
	releaseData, err := g.NewReleaseData(releases)
	if err != nil {
		return nil, err
	}
	return releaseData, nil
}

type ReleaseData struct {
	Releases  map[Version]*VersionData
	generator *MinorReleaseGroupedChangelogGenerator
}

func (g *MinorReleaseGroupedChangelogGenerator) NewReleaseData(releases []*github.RepositoryRelease) (*ReleaseData, error) {
	r := &ReleaseData{
		Releases: make(map[Version]*VersionData),
	}
	for _, release := range releases {
		tag, err := ParseVersion(release.GetTagName())
		if err != nil {
			return nil, UnableToParseVersionError(err, release.GetTagName())
		}


		releaseVersion := &Version{Major: tag.Major, Minor: tag.Minor}
		if r.Releases[*releaseVersion] == nil {
			if err != nil {
				return nil, err
			}
			r.Releases[*releaseVersion] = g.NewVersionData()
		}
		notes, err := g.NewChangelogNotes(release)
		currRelease := r.Releases[*releaseVersion]
		currRelease.ChangelogNotes[*tag] = notes
	}
	return r, nil
}

func (r *ReleaseData) GetReleasesSorted() []Version {
	var versions []Version
	for _, versionData := range r.Releases {
		for version, _ := range versionData.ChangelogNotes {
			versions = append(versions, version)
		}
	}
	SortReleaseVersions(versions)
	return versions
}

func (r *ReleaseData) GetChangelogNotes(v Version) *ChangelogNotes {
	if r == nil || r.Releases == nil {
		return nil
	}
	release, ok := r.Releases[Version{Major: v.Major, Minor: v.Minor}]
	if !ok {
		return nil
	}
	if release.ChangelogNotes == nil {
		return nil
	}
	return release.ChangelogNotes[v]

}

func (r *ReleaseData) String() string {
	var versions []Version
	var b strings.Builder
	for ver := range r.Releases {
		versions = append(versions, ver)
	}
	SortReleaseVersions(versions)
	for _, ver := range versions {
		b.WriteString(H3(ver.String()))
		b.WriteString(r.Releases[ver].String())
	}
	return b.String()
}

func (r *ReleaseData) Dump() (string, error) {
	var versions []Version
	var b strings.Builder
	for ver := range r.Releases {
		versions = append(versions, ver)
	}
	SortReleaseVersions(versions)
	b.WriteRune('[')
	for i, ver := range versions {
		b.WriteString(fmt.Sprintf("{\"%s\":", ver.String()))
		cNotes, err := r.Releases[ver].Dump()
		if err != nil {
			return "", nil
		}
		b.WriteString(cNotes)
		b.WriteRune('}')
		if i != len(versions)-1 {
			b.WriteRune(',')
		}
	}
	b.WriteRune(']')
	return b.String(), nil
}

type VersionData struct {
	ChangelogNotes map[Version]*ChangelogNotes
	generator      *MinorReleaseGroupedChangelogGenerator
}

func (g *MinorReleaseGroupedChangelogGenerator) NewVersionData() *VersionData {
	return &VersionData{
		ChangelogNotes: map[Version]*ChangelogNotes{},
		generator:      g,
	}
}

func (v *VersionData) String() string {
	var versions []Version
	var b strings.Builder
	for ver := range v.ChangelogNotes {
		versions = append(versions, ver)
	}
	SortReleaseVersions(versions)
	for _, ver := range versions {
		changelogNotes := v.ChangelogNotes[ver]
		b.WriteString(H4(getGithubReleaseMarkdownLink(ver.String(), v.generator.RepoOwner, v.generator.Repo) + changelogNotes.HeaderSuffix))
		b.WriteString(changelogNotes.String())
	}
	return b.String()
}

func (v *VersionData) Dump() (string, error) {
	var versions []Version
	var b strings.Builder
	for ver := range v.ChangelogNotes {
		versions = append(versions, ver)
	}
	SortReleaseVersions(versions)
	b.WriteRune('[')
	for i, ver := range versions {
		b.WriteString(fmt.Sprintf("{\"%s\":", ver.String()))
		cNotes, err := v.ChangelogNotes[ver].Dump()
		if err != nil {
			return "", nil
		}
		b.WriteString(cNotes)
		b.WriteRune('}')
		if i != len(versions)-1 {
			b.WriteRune(',')
		}
	}
	b.WriteRune(']')
	return b.String(), nil
}

type ChangelogNotes struct {
	Categories   map[string][]string
	ExtraNotes   []string
	HeaderSuffix string
	CreatedAt int64
}

func NewChangelogNotes() *ChangelogNotes {
	return &ChangelogNotes{Categories: make(map[string][]string)}
}

func (g *MinorReleaseGroupedChangelogGenerator) NewChangelogNotes(r *github.RepositoryRelease) (*ChangelogNotes, error) {
	body := r.GetBody()
	extraNotes, releaseNotes, err := ParseReleaseBody(body)
	if err != nil {
		return nil, err
	}
	return &ChangelogNotes{
		Categories: releaseNotes,
		ExtraNotes: extraNotes,
		CreatedAt: r.GetCreatedAt().Unix(),
	}, nil
}

func (c *ChangelogNotes) String() string {
	var b strings.Builder
	var keys []string
	for k := range c.Categories {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, header := range keys {
		b.WriteString(H5(header))
		for _, note := range c.Categories[header] {
			b.WriteString(UnorderedListItem(note))
		}
	}
	if len(c.ExtraNotes) != 0 {
		b.WriteString(H5("Notes"))
		for _, note := range c.ExtraNotes {
			b.WriteString(UnorderedListItem(note))
		}
	}
	return b.String()
}

func (c *ChangelogNotes) Dump() (string, error) {
	res, err := json.Marshal(c)
	r := strings.NewReplacer("\\r", "", "\\n", "")
	return r.Replace(string(res)), err
}

func (c *ChangelogNotes) Add(other *ChangelogNotes) {
	c.AddWithPrefix(other, "")
}

func (c *ChangelogNotes) AddWithPrefix(other *ChangelogNotes, prefix string) {
	for header, notes := range other.Categories {
		for _, note := range notes {
			c.Categories[header] = append(c.Categories[header], prefix+note)
		}
	}
}

func (c *ChangelogNotes) AddWithPrefixIncludeExtraNotes(other *ChangelogNotes, prefix string) {
	c.AddWithPrefix(other, prefix)
	for _, note := range other.ExtraNotes {
		c.ExtraNotes = append(c.ExtraNotes, prefix+note)
	}
}

func ParseReleaseBody(body string) ([]string, map[string][]string, error) {
	var (
		currentHeader string
		extraNotes    []string
	)
	releaseNotes := make(map[string][]string)
	buf := []byte(body)
	root := goldmark.DefaultParser().Parse(text.NewReader(buf))
	// Translate list of markdown "blocks" to a map of headers to notes
	for n := root.FirstChild(); n != nil; n = n.NextSibling() {
		switch typedNode := n.(type) {
		case *ast.Paragraph:
			{
				if child := typedNode.FirstChild(); child.Kind() == ast.KindEmphasis {
					if child.(*ast.Emphasis).Level == 2 {
						// Header
						currentHeader = string(typedNode.Text(buf))
						continue
					}
				}
				// This section will handles any paragraphs that do not show up under headers e.g. "This release build failed"
				v := typedNode.Lines().At(0)
				note := string(v.Value(buf))
				if currentHeader != "" {
					releaseNotes[currentHeader] = append(releaseNotes[currentHeader], note)
				} else {
					//any extra text e.g. "This release build has failed", only used for enterprise release notes
					extraNotes = append(extraNotes, note)
				}
			}
		case *ast.List:
			{
				// Only add release notes if we are under a current header
				for child := n.FirstChild(); child != nil; child = child.NextSibling() {
					v := child.FirstChild().Lines().At(0)
					releaseNote := string(v.Value(buf))
					if currentHeader != "" {
						releaseNotes[currentHeader] = append(releaseNotes[currentHeader], releaseNote)
					} else {
						//any extra text that may be in a list but not under a heading
						extraNotes = append(extraNotes, releaseNote)
					}
				}
			}
		default:
			{
				return nil, nil, fmt.Errorf("unable to parse %s", n.Text(buf))
			}
		}
	}
	return extraNotes, releaseNotes, nil
}

// Sorts a slice of versions in descending order by version e.g. v1.6.1, v1.6.0, v1.6.0-beta9
func SortReleaseVersions(versions []Version) {
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].MustIsGreaterThanOrEqualTo(versions[j])
	})
}

func getGithubReleaseMarkdownLink(tag, repoOwner, repo string) string {
	link := fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", repoOwner, repo, tag)
	return Link(tag, link)
}
