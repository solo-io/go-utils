package changelogdocutils

import (
	"context"
	"fmt"
	"github.com/google/go-github/v32/github"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/go-utils/versionutils"
	"io/ioutil"
	"net/http"
	"regexp"
)

type MergedReleaseGenerator struct {
	client               *github.Client
	repoOwner            string
	enterpriseRepo       string
	openSourceRepo       string
	releaseDepMap        map[Version]*Version
	findDependentVersion func(*Version, map[Version]*Version) (*Version, error)
}

func FindDependentVersionFn(enterpriseVersion *Version, cache map[Version]*Version) (*Version, error) {
	if cache != nil {
		if dep := cache[*enterpriseVersion]; dep != nil {
			return dep, nil
		}
	}
	versionTag := enterpriseVersion.String()
	dependencyUrl := fmt.Sprintf("https://storage.googleapis.com/gloo-ee-dependencies/%s/dependencies", versionTag[1:])
	request, err := http.NewRequest("GET", dependencyUrl, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(`.*gloo.*(v.*)`)
	if err != nil {
		return nil, err
	}
	matches := re.FindStringSubmatch(string(body))
	if len(matches) != 2 {
		return nil, eris.Errorf("unable to get gloo dependency for gloo enterprise version %s\n response from google storage API: %s", versionTag, string(body))
	}
	glooVersionTag := matches[1]
	version, err := ParseVersion(glooVersionTag)
	if err != nil {
		return nil, err
	}
	if cache != nil{
		cache[*enterpriseVersion] = version
	}
	return version, nil
}

func NewMergedReleaseGenerator(client *github.Client, repoOwner, enterpriseRepo, ossRepo string) *MergedReleaseGenerator {
	return &MergedReleaseGenerator{
		client:               client,
		repoOwner:            repoOwner,
		enterpriseRepo:       enterpriseRepo,
		openSourceRepo:       ossRepo,
		releaseDepMap:        map[Version]*Version{},
		findDependentVersion: FindDependentVersionFn,
	}
}

func (g *MergedReleaseGenerator) Generate(ctx context.Context) (string, error) {
	ossReleases, err := NewMinorReleaseGroupedChangelogGenerator(g.client, g.repoOwner, g.openSourceRepo).
		GetReleaseData(ctx)
	if err != nil {
		return "", err
	}
	enterpriseReleases, err := NewMinorReleaseGroupedChangelogGenerator(g.client, g.repoOwner, g.enterpriseRepo).
		GetReleaseData(ctx)
	if err != nil {
		return "", err
	}
	// Get releases from max version to min version (e.g. 1.8.0, 1.8.0-beta2, 1.8.0-beta1...)
	enterpriseSorted := enterpriseReleases.GetReleasesSorted()
	osSorted := ossReleases.GetReleasesSorted()
	for _, release := range enterpriseSorted {
		// Build out release dependency map (enterprise -> oss)
		_, err = g.findDependentVersion(&release, g.releaseDepMap)
		if err != nil {
			continue
			//return "", err
		}
	}

	for idx, eRelease := range enterpriseSorted {
		if idx >= len(enterpriseSorted)-1 {
			break
		}
		earlierVersion := enterpriseSorted[idx+1]
		// If earlier version doesn't have a OSS dependency, look for next version that does to compute changelog diff
		for i := 1;
			g.releaseDepMap[earlierVersion] == nil && idx+i < len(enterpriseSorted)-1;
		i, earlierVersion = i+1, enterpriseSorted[idx+i] {
		}
		// Find all open source released versions between consecutive enterprise versions
		earlierDep, currentDep := g.releaseDepMap[earlierVersion], g.releaseDepMap[eRelease]
		if earlierDep != nil && currentDep != nil {
			depVersions := GetOtherRepoDepsBetweenVersions(osSorted, earlierDep, currentDep)
			var finalChangelogNotes = NewChangelogNotes()
			for _, version := range depVersions {
				prefix := fmt.Sprintf("(From OSS %s) ", getGithubReleaseMarkdownLink(version.String(), g.repoOwner, g.openSourceRepo))
				finalChangelogNotes.AddWithPrefix(ossReleases.GetChangelogNotes(version), prefix)
			}
			minorReleaseChangelogNotes := enterpriseReleases.GetChangelogNotes(eRelease)
			minorReleaseChangelogNotes.HeaderSuffix = fmt.Sprintf(" (Uses OSS %s)", getGithubReleaseMarkdownLink(currentDep.String(), g.repoOwner, g.openSourceRepo))
			minorReleaseChangelogNotes.Add(finalChangelogNotes)
		}
	}
	res, err := enterpriseReleases.Dump()
	return res, err
}

func GetOtherRepoDepsBetweenVersions(otherRepoReleasesSorted []Version, earlierVersion, laterVersion *Version) []Version {
	var i, j = -1, -1
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
	if i < 0 || j < 0 || j < i {
		return nil
	}
	return otherRepoReleasesSorted[i:j]
}
