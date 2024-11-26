package changelogdocutils

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/go-git/go-git/v5"
	http2 "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-github/v32/github"
	"github.com/solo-io/go-utils/githubutils"
	. "github.com/solo-io/go-utils/versionutils"
)

type DependencyFn func(*Version) (*Version, error)

type Options struct {
	// Github user/org
	RepoOwner,
	// Only repo used in MinorReleaseGenerator
	// This is the enterprise repo used in MergedReleaseGenerator
	MainRepo,
	// Unused in MinorReleaseGenerator
	// This is the Open Source repo in MergedReleaseGenerator
	DependentRepo string
	// List of Github Repository Releases for the Main Repo
	MainRepoReleases []*github.RepositoryRelease
	// List of Github Repository Releases for the Dependent Repo
	DependentRepoReleases []*github.RepositoryRelease

	// NumVersions sets the maximum amount of releases to be fetched from Github.
	// Once fetched, MaxVersion and MinVersion are bounds for which versions are included
	// in the output. The following three only bound the
	NumVersions int
	// Maximum release version to be included in this changelog. If not specified, all releases >= MinVersion
	// will be included
	MaxVersion *Version
	// Minimum release version to display on this changelog. If not specified, all releases <= MaxVersion
	// will be included
	MinVersion *Version
}

type MergedReleaseGenerator struct {
	client         *github.Client
	releaseDepMap  map[Version]*Version
	opts           Options
	dependencyFunc DependencyFn
}

func NewMergedReleaseGenerator(opts Options, client *github.Client) *MergedReleaseGenerator {
	generator := &MergedReleaseGenerator{
		opts:          opts,
		client:        client,
		releaseDepMap: map[Version]*Version{},
	}
	generator.dependencyFunc = generator.GetOpenSourceDependency
	return generator
}

func NewMergedReleaseGeneratorWithDepFn(opts Options, client *github.Client, depFn DependencyFn) *MergedReleaseGenerator {
	gen := NewMergedReleaseGenerator(opts, client)
	gen.dependencyFunc = depFn
	return gen
}

/*
The merged release generator has 3 steps:
1. Fetches enterprise repo release notes
2. Gets open source dependency for each enterprise version
3. Merges open source release notes into enterprise version release notes
*/
func (g *MergedReleaseGenerator) GenerateJSON(ctx context.Context) (string, error) {
	var err error
	enterpriseReleases, err := g.GetMergedEnterpriseRelease(ctx)
	if err != nil {
		return "", err
	}
	// Coerce LabelVersion to 1 for alpine labels. Without this, there will be duplicate versions
	// in the changelog
	for mainRelease, mainVersionData := range enterpriseReleases.Releases {
		for version, changeLogNote := range mainVersionData.ChangelogNotes {
			if version.Label == "alpine" && version.LabelVersion == 0 {
				clonedChangeLogNotes := NewChangelogNotes()
				clonedChangeLogNotes.Add(changeLogNote)
				newVer := Version{
					Major:        version.Major,
					Minor:        version.Minor,
					Patch:        version.Patch,
					Label:        version.Label,
					LabelVersion: 1,
				}
				enterpriseReleases.Releases[mainRelease].ChangelogNotes[newVer] = changeLogNote
				delete(enterpriseReleases.Releases[mainRelease].ChangelogNotes, version)
			}
		}
	}
	var out struct {
		Opts        Options
		ReleaseData *ReleaseData
	}
	out.Opts = Options{
		RepoOwner:     g.opts.RepoOwner,
		MainRepo:      g.opts.MainRepo,
		DependentRepo: g.opts.DependentRepo,
	}
	out.ReleaseData = enterpriseReleases
	res, err := json.Marshal(out)
	return string(res), err
}

func (g *MergedReleaseGenerator) GetMergedEnterpriseRelease(ctx context.Context) (*ReleaseData, error) {

	enterpriseReleases, err := NewMinorReleaseGroupedChangelogGenerator(g.opts, g.client).
		GetReleaseData(ctx, g.opts.MainRepoReleases)
	if err != nil {
		return nil, err
	}
	ossOpts := g.opts
	ossOpts.MainRepo = g.opts.DependentRepo
	ossReleases, err := NewMinorReleaseGroupedChangelogGenerator(ossOpts, g.client).
		GetReleaseData(ctx, g.opts.DependentRepoReleases)
	if err != nil {
		return nil, err
	}
	return g.MergeEnterpriseReleaseWithOS(enterpriseReleases, ossReleases)
}

func (g *MergedReleaseGenerator) MergeEnterpriseReleaseWithOS(enterpriseReleases, osReleases *ReleaseData) (*ReleaseData, error) {
	// Get openSourceReleases and enterpriseReleases sorted from max version to min version
	// (e.g. 1.8.0, 1.8.0-beta2, 1.8.0-beta1, 1.7.15, 1.7.14...)
	enterpriseSorted := enterpriseReleases.GetReleasesSorted()
	osSorted := osReleases.GetReleasesSorted()
	for _, release := range enterpriseSorted {
		// Build out release dependency map (enterprise -> oss) using releaseDepMap as cache
		if g.releaseDepMap[release] == nil {
			var (
				dep *Version
				err error
			)
			dep, err = g.dependencyFunc(&release)
			if err != nil {
				continue
			}
			g.releaseDepMap[release] = dep
		}
	}

	for idx, eRelease := range enterpriseSorted {
		var earlierVersion, earlierOSDep, currentOSDep *Version
		if idx < len(enterpriseSorted)-1 {
			earlierVersion = &enterpriseSorted[idx+1]
		}
		// If earlier version doesn't have an OSS dependency, look for next version that does to compute changelog diff
		if earlierVersion != nil {
			for i := 1; g.releaseDepMap[*earlierVersion] == nil && idx+i < len(enterpriseSorted)-1; i, earlierVersion = i+1, &enterpriseSorted[idx+i] {
			}
			earlierOSDep = g.releaseDepMap[*earlierVersion]
		}
		// Find all open source released versions between consecutive enterprise versions
		currentOSDep = g.releaseDepMap[eRelease]
		if currentOSDep != nil {
			depVersions := GetOtherRepoDepsBetweenVersions(osSorted, earlierOSDep, currentOSDep)
			var finalChangelogNotes = NewChangelogNotes()
			for _, version := range depVersions {
				//prefix := fmt.Sprintf("(From OSS %s) ", getGithubReleaseMarkdownLink(version.String(), g.RepoOwner, g.openSourceRepo))
				notes, err := osReleases.GetChangelogNotes(version)
				if err != nil {
					return nil, err
				}
				finalChangelogNotes.AddWithDependentVersion(notes, version)
			}
			minorReleaseChangelogNotes, err := enterpriseReleases.GetChangelogNotes(eRelease)
			if err != nil {
				return nil, err
			}
			minorReleaseChangelogNotes.HeaderSuffix = fmt.Sprintf(" (Uses OSS %s)", getGithubReleaseMarkdownLink(currentOSDep.String(), g.opts.RepoOwner, g.opts.DependentRepo))
			minorReleaseChangelogNotes.Add(finalChangelogNotes)
		}
	}
	return enterpriseReleases, nil
}

func GetOtherRepoDepsBetweenVersions(otherRepoReleasesSorted []Version, earlierVersion, laterVersion *Version) []Version {
	if earlierVersion == nil {
		return []Version{*laterVersion}
	}
	var i, j = -1, -1
	// otherRepoReleasesSorted is sorted from highest semver to lowest semver v1.7.0 -> v1.0.0
	for idx, release := range otherRepoReleasesSorted {
		if release == *laterVersion {
			i = idx
		}
		if release == *earlierVersion {
			j = idx
			break
		}
		// Don't look for dependent versions across major / minor versions
		if i != -1 && (release.Major != earlierVersion.Major || release.Minor != earlierVersion.Minor) {
			j = idx
			break
		}
	}
	if i < 0 || j < 0 || j+1 < i {
		return nil
	}
	return otherRepoReleasesSorted[i:j]
}

func getGithubReleaseMarkdownLink(tag, repoOwner, repo string) string {
	return fmt.Sprintf("[%s](https://github.com/%s/%s/releases/tag/%s)", tag, repoOwner, repo, tag)
}

// Looks for
// repoOwner/otherRepo v1.x.x-x
// in go.mod string (goMod) and returns the package version
func getPkgVersionFromGoMod(goMod, repoOwner, repo string) (*Version, error) {
	semverRegex := fmt.Sprintf("%s/%s\\s+(v((([0-9]+)\\.([0-9]+)\\.([0-9]+)(?:-([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?)(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?))",
		repoOwner, repo)
	regex := regexp.MustCompile(semverRegex)
	// Find version of open source dependency
	matches := regex.FindStringSubmatch(goMod)
	if len(matches) < 2 {
		return nil, fmt.Errorf("unable to find dependency")
	}
	return ParseVersion(matches[1])
}

/*
Enterprise changelogs have the option to "merge" open source changelogs. The following function retrieves and returns
the open source version that an enterprise version depends on. It checks out the enterprise go.mod at the release version
and looks for the open source dependency ({RepoOwner}/{DependentRepo}).
*/
func (g *MergedReleaseGenerator) GetOpenSourceDependency(enterpriseVersion *Version) (*Version, error) {
	files, err := githubutils.GetFilesFromGit(context.TODO(), g.client, g.opts.RepoOwner, g.opts.MainRepo, enterpriseVersion.String(), "go.mod")
	if err != nil {
		return nil, fmt.Errorf("error fetching dependency for enterprise version %s: %s", enterpriseVersion.String(), err.Error())
	}
	if len(files) < 1 {
		return nil, fmt.Errorf("unable to find go.mod file in enteprise repository")
	}
	content, err := files[0].GetContent()
	if err != nil {
		return nil, err
	}

	return getPkgVersionFromGoMod(content, g.opts.RepoOwner, g.opts.DependentRepo)
}

// HOF that returns a Dependency function. The returned function generates
// changelogs faster than the default but requires passing in the github token.
func GetOSDependencyFunc(repoOwner, enterpriseRepo, osRepo, githubToken string) (DependencyFn, error) {
	// Clones repo in-memory, we only want to clone the repo into the variable once, hence the HOF to
	// pass the repo variable in the function's closure.
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: fmt.Sprintf("https://github.com/%s/%s", repoOwner, enterpriseRepo),
		Auth: &http2.BasicAuth{
			Username: "nonEmptyString", // https://github.com/go-git/go-git/blob/master/_examples/clone/auth/basic/access_token/main.go#L24
			Password: githubToken,
		},
	})
	if err != nil {
		return nil, err
	}
	dependencyFn := func(v *Version) (*Version, error) {
		tagRef, err := repo.Tag(v.String())
		if err != nil {
			return nil, err
		}
		commit, err := repo.CommitObject(tagRef.Hash())
		if err != nil {
			return nil, err
		}
		gomod, err := commit.File("go.mod")
		if err != nil {
			return nil, err
		}
		content, err := gomod.Contents()
		if err != nil {
			return nil, err
		}
		return getPkgVersionFromGoMod(content, repoOwner, osRepo)
	}
	return dependencyFn, nil
}
