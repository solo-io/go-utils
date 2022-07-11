package securityscanutils

import (
	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v32/github"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/go-utils/versionutils"
)

// The securityScanRepositoryReleasePredicate is responsible for defining which
// github.RepositoryRelease artifacts should be included in the bulk security scan
// At the moment, the two requirements are that:
// 1. The release is not a pre-release or draft
// 2. The release matches the configured version constraint
type securityScanRepositoryReleasePredicate struct {
	versionConstraint *semver.Constraints
}

func NewSecurityScanRepositoryReleasePredicate(constraint *semver.Constraints) *securityScanRepositoryReleasePredicate {
	return &securityScanRepositoryReleasePredicate{
		versionConstraint: constraint,
	}
}

func (s *securityScanRepositoryReleasePredicate) Apply(release *github.RepositoryRelease) bool {
	if release.GetPrerelease() || release.GetDraft() {
		// Do not include pre-releases or drafts
		return false
	}

	versionToTest, err := semver.NewVersion(release.GetTagName())
	if err != nil {
		// If we are unable to parse the release version, we do not include it in the filtered list
		log.Warnf("unable to parse release version %s", release.GetTagName())
		return false
	}

	if !s.versionConstraint.Check(versionToTest) {
		// If the release version does not pass the version constraint, do not include
		return false
	}

	// If all checks have passed, include the release
	return true
}

// The latestPatchRepositoryReleasePredicate returns true if a provided RepositoryRelease
// is the latest patch release on a long-term support (LTS) branch
// For example: The latest patch release of v1.0.0, v1.0.1, v1.0.2 is v1.0.2 (https://semver.org/)
type latestPatchRepositoryReleasePredicate struct {
	releasesByTagName map[string]*github.RepositoryRelease
}

func NewLatestPatchRepositoryReleasePredicate(releases []*github.RepositoryRelease) *latestPatchRepositoryReleasePredicate {
	githubutils.SortReleasesBySemver(releases)

	// We could use maxint but we dont really care
	// as we can just check if major minor changed
	latestPatchReleasesByTagName := make(map[string]*github.RepositoryRelease)

	recentMajor := -1
	recentMinor := -1
	for _, release := range releases {
		version, err := versionutils.ParseVersion(release.GetTagName())
		if err != nil {
			continue
		}
		if version.Major == recentMajor && version.Minor == recentMinor {
			continue
		}

		// This is the largest patch release
		recentMajor = version.Major
		recentMinor = version.Minor
		latestPatchReleasesByTagName[release.GetTagName()] = release
	}

	return &latestPatchRepositoryReleasePredicate{
		releasesByTagName: latestPatchReleasesByTagName,
	}
}

func (s *latestPatchRepositoryReleasePredicate) Apply(release *github.RepositoryRelease) bool {
	_, ok := s.releasesByTagName[release.GetTagName()]
	return ok
}
