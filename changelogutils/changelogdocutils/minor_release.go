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
	"math"
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
	opts      Options
}

func NewMinorReleaseGroupedChangelogGenerator(opts Options, client *github.Client) *MinorReleaseGroupedChangelogGenerator {
	if opts.NumVersions == 0{
		opts.NumVersions = math.MaxInt64
	}
	return &MinorReleaseGroupedChangelogGenerator{
		opts:      opts,
		Client:    client,
	}
}

func (g *MinorReleaseGroupedChangelogGenerator) GenerateJSON(ctx context.Context) (string, error) {
	var err error
	releaseData, err := g.GetReleaseData(ctx)
	if err != nil {
		return "", err
	}
	var out struct {
		Opts        Options
		ReleaseData *ReleaseData
	}
	out.Opts = g.opts
	out.ReleaseData = releaseData
	res, err := json.Marshal(out)
	return string(res), err
}

func (g *MinorReleaseGroupedChangelogGenerator) GetReleaseData(ctx context.Context) (*ReleaseData, error) {
	releases, err := githubutils.GetAllRepoReleasesWithMax(ctx, g.Client, g.opts.RepoOwner, g.opts.Repo, g.opts.NumVersions)
	if err != nil {
		return nil, err
	}
	releaseData, err := g.NewReleaseData(releases)
	if err != nil {
		return nil, err
	}
	return releaseData, nil
}

// Release Data is mapped such that it easy to group by minor release
// e.g. Releases will be a map of v1.2.0 -> VersionData, v1.3.0 -> VersionData
// VersionData will contain information for individual Versions
type ReleaseData struct {
	Releases  map[Version]*VersionData
	generator *MinorReleaseGroupedChangelogGenerator
}

// Contains Changelog enterpriseNotes for Individual openSourceReleases
// ChangelogNotes is a map of individiual openSourceReleases to enterpriseNotes
// e.g. v1.2.5-beta3 -> ChangelogNotes, v1.4.0 -> ChangelogNotes
type VersionData struct {
	ChangelogNotes map[Version]*ChangelogNotes
	generator      *MinorReleaseGroupedChangelogGenerator
}

type ChangelogNotes struct {
	Categories   map[string][]*Note
	ExtraNotes   []*Note
	HeaderSuffix string
	CreatedAt    int64
}

func (g *MinorReleaseGroupedChangelogGenerator) NewReleaseData(releases []*github.RepositoryRelease) (*ReleaseData, error) {
	r := &ReleaseData{
		Releases: make(map[Version]*VersionData),
	}
	for _, release := range releases {
		tag, err := ParseVersion(release.GetTagName())
		if err != nil {
			// Release name doesn't follow proper semantic versioning, skip
			continue
		}
		if !g.isBetweenMinAndMaxVersions(*tag){
			continue
		}

		releaseVersion := GetMajorAndMinorVersionPtr(tag)
		if r.Releases[*releaseVersion] == nil {
			if err != nil {
				return nil, err
			}
			r.Releases[*releaseVersion] = g.NewVersionData()
		}
		notes, err := g.NewChangelogNotes(release)
		if err != nil {
			return nil, err
		}
		currRelease := r.Releases[*releaseVersion]
		currRelease.ChangelogNotes[*tag] = notes
	}
	return r, nil
}

func (g *MinorReleaseGroupedChangelogGenerator) isBetweenMinAndMaxVersions(version Version) bool{
	res := true
	if g.opts.MinVersion != nil{
		res = version.MustIsGreaterThanOrEqualTo(*g.opts.MinVersion)
	}
	if g.opts.MaxVersion != nil {
		res = res && !version.MustIsGreaterThanOrEqualTo(*g.opts.MaxVersion)
	}
	return res
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
	release, ok := r.Releases[GetMajorAndMinorVersion(&v)]
	if !ok {
		return nil
	}
	if release.ChangelogNotes == nil {
		return nil
	}
	return release.ChangelogNotes[v]

}

func (r *ReleaseData) MarshalJSON() ([]byte, error) {
	var versions []Version
	var b strings.Builder
	for ver := range r.Releases {
		versions = append(versions, ver)
	}
	SortReleaseVersions(versions)
	b.WriteRune('[')
	for i, ver := range versions {
		b.WriteString(fmt.Sprintf("{\"%s\":", ver.String()))
		cNotes, err := json.Marshal(r.Releases[ver])
		if err != nil {
			return nil, nil
		}
		b.Write(cNotes)
		b.WriteRune('}')
		if i != len(versions)-1 {
			b.WriteRune(',')
		}
	}
	b.WriteRune(']')
	return []byte(b.String()), nil
}

func (g *MinorReleaseGroupedChangelogGenerator) NewVersionData() *VersionData {
	return &VersionData{
		ChangelogNotes: map[Version]*ChangelogNotes{},
		generator:      g,
	}
}

func (v *VersionData) MarshalJSON() ([]byte, error) {
	var versions []Version
	var b strings.Builder
	for ver := range v.ChangelogNotes {
		versions = append(versions, ver)
	}
	SortReleaseVersions(versions)
	b.WriteRune('[')
	for i, ver := range versions {
		b.WriteString(fmt.Sprintf("{\"%s\":", ver.String()))
		cNotes, err := json.Marshal(v.ChangelogNotes[ver])
		if err != nil {
			return nil, nil
		}
		b.Write(cNotes)
		b.WriteRune('}')
		if i != len(versions)-1 {
			b.WriteRune(',')
		}
	}
	b.WriteRune(']')
	return []byte(b.String()), nil
}

func NewChangelogNotes() *ChangelogNotes {
	return &ChangelogNotes{Categories: make(map[string][]*Note)}
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
		CreatedAt:  r.GetCreatedAt().Unix(),
	}, nil
}

func (c *ChangelogNotes) Dump() (string, error) {
	res, err := json.Marshal(c)
	r := strings.NewReplacer("\\r", "", "\\n", "")
	return r.Replace(string(res)), err
}

func (c *ChangelogNotes) Add(other *ChangelogNotes) {
	for header, notes := range other.Categories {
		for _, note := range notes {
			c.Categories[header] = append(c.Categories[header], note)
		}
	}
}

func (c *ChangelogNotes) AddWithDependentVersion(other *ChangelogNotes, depVersion Version) {
	for header, notes := range other.Categories {
		for _, note := range notes {
			c.Categories[header] = append(c.Categories[header], &Note{note.Note, &depVersion})
		}
	}
}

func (c *ChangelogNotes) AddWithDependentVersionIncludeExtraNotes(other *ChangelogNotes, depVersion Version) {
	c.AddWithDependentVersion(other, depVersion)
	for _, note := range other.ExtraNotes {
		c.ExtraNotes = append(c.ExtraNotes, &Note{note.Note, &depVersion})
	}
}

type Note struct {
	Note string
	// Indicates which version of the dependent Repo that this note is from
	FromDependentVersion *Version
}

func (c *Note) MarshalJSON() ([]byte, error) {
	note, err := json.Marshal(c.Note)
	if err != nil {
		return nil, err
	}
	str := fmt.Sprintf(`{"Note": %s`, note)
	if c.FromDependentVersion != nil {
		version, err := json.Marshal(c.FromDependentVersion.String())
		if err != nil {
			return nil, err
		}
		str += fmt.Sprintf(`, "FromDependentVersion":%s`, version)
	}
	str += "}"
	return []byte(str), nil
}

func ParseReleaseBody(body string) ([]*Note, map[string][]*Note, error) {
	var (
		currentHeader string
		extraNotes    []*Note
	)
	releaseNotes := make(map[string][]*Note)
	buf := []byte(body)
	root := goldmark.DefaultParser().Parse(text.NewReader(buf))
	// Translate list of markdown "blocks" to a map of headers to enterpriseNotes
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
					releaseNotes[currentHeader] = append(releaseNotes[currentHeader], &Note{Note: note})
				} else {
					//any extra text e.g. "This release build has failed", only used for enterprise release enterpriseNotes
					extraNotes = append(extraNotes, &Note{Note: note})
				}
			}
		case *ast.List:
			{
				// Only add release enterpriseNotes if we are under a current header
				for child := n.FirstChild(); child != nil; child = child.NextSibling() {
					if child.FirstChild().Lines().Len() > 0 {
						v := child.FirstChild().Lines().At(0)
						releaseNote := string(v.Value(buf))
						if currentHeader != "" {
							releaseNotes[currentHeader] = append(releaseNotes[currentHeader], &Note{Note: releaseNote})
						} else {
							//any extra text that may be in a list but not under a heading
							extraNotes = append(extraNotes, &Note{Note: releaseNote})
						}
					}
				}
			}
		default:
			{
				continue
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

func GetMajorAndMinorVersion(v *Version) Version {
	return Version{Major: v.Major, Minor: v.Minor}
}

func GetMajorAndMinorVersionPtr(v *Version) *Version {
	return &Version{Major: v.Major, Minor: v.Minor}
}
