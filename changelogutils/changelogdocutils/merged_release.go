package changelogdocutils

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/google/go-github/v32/github"
    "github.com/solo-io/go-utils/githubutils"
    . "github.com/solo-io/go-utils/versionutils"
    "regexp"
)

type DependencyFn func(*Version) (*Version, error)

type Options struct {
    // Number of versions from the most previous version to display in changelog
    NumVersions int
    // Maximum version changelog to display on this changelog
    MaxVersion *Version
    // Minimum version changelog to display on this changelog
    MinVersion *Version
    ProjectName,
    RepoOwner,
    MainRepo,
    DependentRepo string
}

type MergedReleaseGenerator struct {
    client         *github.Client
    releaseDepMap  map[Version]*Version
    opts           Options
    DependencyFunc DependencyFn
}

func NewMergedReleaseGenerator(opts Options, client *github.Client) *MergedReleaseGenerator {
    return &MergedReleaseGenerator{
        opts:          opts,
        client:        client,
        releaseDepMap: map[Version]*Version{},
    }
}

func NewMergedReleaseGeneratorWithDepFn(opts Options, client *github.Client, depFn DependencyFn) *MergedReleaseGenerator {
	gen :=  NewMergedReleaseGenerator(opts, client)
	gen.DependencyFunc = depFn
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
    var out struct {
        Opts        Options
        ReleaseData *ReleaseData
    }
    out.Opts = g.opts
    out.ReleaseData = enterpriseReleases
    res, err := json.Marshal(out)
    return string(res), err
}

func (g *MergedReleaseGenerator) GetMergedEnterpriseRelease(ctx context.Context) (*ReleaseData, error) {
    ossReleases, err := NewMinorReleaseGroupedChangelogGenerator(Options{
        RepoOwner: "solo-io",
        MainRepo:  g.opts.DependentRepo,
    }, g.client).
        GetReleaseData(ctx)
    if err != nil {
        return nil, err
    }
    enterpriseReleases, err := NewMinorReleaseGroupedChangelogGenerator(g.opts, g.client).
        GetReleaseData(ctx)
    if err != nil {
        return nil, err
    }
    return g.MergeEnterpriseReleaseWithOS(ctx, enterpriseReleases, ossReleases)
}

func (g *MergedReleaseGenerator) MergeEnterpriseReleaseWithOS(ctx context.Context, enterpriseReleases, osReleases *ReleaseData) (*ReleaseData, error) {
    // Get openSourceReleases from max version to min version (e.g. 1.8.0, 1.8.0-beta2, 1.8.0-beta1...)
    enterpriseSorted := enterpriseReleases.GetReleasesSorted()
    osSorted := osReleases.GetReleasesSorted()
    for _, release := range enterpriseSorted {
        // Build out release dependency map (enterprise -> oss) using releaseDepMap as cache
        if g.releaseDepMap[release] == nil {
            var (
                dep *Version
                err error
            )
            if g.DependencyFunc != nil {
                dep, err = g.DependencyFunc(&release)
            } else {
                dep, err = g.GetOpenSourceDependency(&release)
            }
            if err != nil {
                continue
                //return "", err
            }
            g.releaseDepMap[release] = dep
        }
    }

    for idx, eRelease := range enterpriseSorted {
        var earlierVersion, earlierOSDep, currentOSDep *Version
        if idx < len(enterpriseSorted)-1 {
            earlierVersion = &enterpriseSorted[idx+1]
        }
        // If earlier version doesn't have a OSS dependency, look for next version that does to compute changelog diff
        if earlierVersion != nil {
            for i := 1;
                g.releaseDepMap[*earlierVersion] == nil && idx+i < len(enterpriseSorted)-1;
            i, earlierVersion = i+1, &enterpriseSorted[idx+i] {
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
                finalChangelogNotes.AddWithDependentVersion(osReleases.GetChangelogNotes(version), version)
            }
            minorReleaseChangelogNotes := enterpriseReleases.GetChangelogNotes(eRelease)
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
    // Looks for
    // ...repoOwner/otherRepo v1.x.x-x
    // and captures semantic version
    semverRegex := fmt.Sprintf("%s/%s\\s+(v((([0-9]+)\\.([0-9]+)\\.([0-9]+)(?:-([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?)(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?))",
        g.opts.RepoOwner, g.opts.DependentRepo)
    regex := regexp.MustCompile(semverRegex)
    // Find version of open source dependency
    matches := regex.FindStringSubmatch(content)
    if len(matches) < 2 {
        return nil, fmt.Errorf("unable to find version %s dependency", enterpriseVersion.String())
    }
    return ParseVersion(matches[1])
}
