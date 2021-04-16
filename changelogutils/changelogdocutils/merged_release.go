package changelogdocutils

import (
	"context"
	"fmt"
	"github.com/google/go-github/v32/github"
	. "github.com/solo-io/go-utils/versionutils"
)

type DependencyFn func(*Version) (*Version, error)

type MergedReleaseGeneratorOptions struct {
	MaxVersions int
}

type MergedReleaseGenerator struct {
	client               *github.Client
	repoOwner            string
	enterpriseRepo       string
	openSourceRepo       string
	releaseDepMap        map[Version]*Version
	FindDependentVersion DependencyFn
	opts                 MergedReleaseGeneratorOptions
}

func NewMergedReleaseGenerator(opts MergedReleaseGeneratorOptions, client *github.Client, repoOwner, enterpriseRepo, ossRepo string, dependencyFunc DependencyFn) *MergedReleaseGenerator {
	return &MergedReleaseGenerator{
		opts:                 opts,
		client:               client,
		repoOwner:            repoOwner,
		enterpriseRepo:       enterpriseRepo,
		openSourceRepo:       ossRepo,
		releaseDepMap:        map[Version]*Version{},
		FindDependentVersion: dependencyFunc,
	}

}

func (g *MergedReleaseGenerator) Generate(ctx context.Context) (string, error) {
	enterpriseReleases, err := g.GetMergedEnterpriseRelease(ctx)
	if err != nil {
		return "", err
	}
	return enterpriseReleases.String(), nil
}

func (g *MergedReleaseGenerator) GenerateJSON(ctx context.Context) (string, error) {
	enterpriseReleases, err := g.GetMergedEnterpriseRelease(ctx)
	if err != nil {
		return "", err
	}
	return enterpriseReleases.Dump()
}

func (g *MergedReleaseGenerator) GetMergedEnterpriseRelease(ctx context.Context) (*ReleaseData, error) {
	ossReleases, err := NewMinorReleaseGroupedChangelogGenerator(MinorReleaseOpts{}, g.client, g.repoOwner, g.openSourceRepo).
		GetReleaseData(ctx)
	if err != nil {
		return nil, err
	}
	opts := MinorReleaseOpts{
		MaxVersions: g.opts.MaxVersions,
	}
	enterpriseReleases, err := NewMinorReleaseGroupedChangelogGenerator(opts, g.client, g.repoOwner, g.enterpriseRepo).
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
			dep, err := g.FindDependentVersion(&release)
			if err != nil {
				continue
				//return "", err
			}
			g.releaseDepMap[release] = dep
		}
	}

	for idx, eRelease := range enterpriseSorted {
		var earlierVersion, earlierDep, currentDep *Version
		if idx < len(enterpriseSorted)-1 {
			earlierVersion = &enterpriseSorted[idx+1]
		}
		// If earlier version doesn't have a OSS dependency, look for next version that does to compute changelog diff
		if earlierVersion != nil {
			for i := 1;
				g.releaseDepMap[*earlierVersion] == nil && idx+i < len(enterpriseSorted)-1;
			i, earlierVersion = i+1, &enterpriseSorted[idx+i] {
			}
			earlierDep = g.releaseDepMap[*earlierVersion]
		}
		// Find all open source released versions between consecutive enterprise versions
		currentDep = g.releaseDepMap[eRelease]
		if currentDep != nil {
			depVersions := GetOtherRepoDepsBetweenVersions(osSorted, earlierDep, currentDep)
			var finalChangelogNotes = NewChangelogNotes()
			for _, version := range depVersions {
				//prefix := fmt.Sprintf("(From OSS %s) ", getGithubReleaseMarkdownLink(version.String(), g.repoOwner, g.openSourceRepo))
				finalChangelogNotes.AddWithDependentVersion(osReleases.GetChangelogNotes(version), version)
			}
			minorReleaseChangelogNotes := enterpriseReleases.GetChangelogNotes(eRelease)
			minorReleaseChangelogNotes.HeaderSuffix = fmt.Sprintf(" (Uses OSS %s)", getGithubReleaseMarkdownLink(currentDep.String(), g.repoOwner, g.openSourceRepo))
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
	return otherRepoReleasesSorted[i : j+1]
}
