package securityscanutils

import (
	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v32/github"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/go-utils/versionutils"
)

// The SecurityScanRepositoryReleasePredicate is responsible for defining which
// github.RepositoryRelease artifacts should be included in the bulk security scan
// At the moment, the two requirements are that:
// 1. The release is not a pre-release or draft
// 2. The release matches the configured version constraint
type SecurityScanRepositoryReleasePredicate struct {
	versionConstraint *semver.Constraints
}

func (s *SecurityScanRepositoryReleasePredicate) Apply(release *github.RepositoryRelease) bool {
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

type LTSOnlyRepositoryReleasePredicate struct {
	releaseSet []*github.RepositoryRelease
}

func NewLTSOnlyRepositoryReleasePredicate(releases []*github.RepositoryRelease) *LTSOnlyRepositoryReleasePredicate {
	// We could use maxint but we dont really care
	// as we can just check if major minor changed
	var ltsOnlyReleases []*github.RepositoryRelease

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
		ltsOnlyReleases = append(ltsOnlyReleases, release)
	}

	return &LTSOnlyRepositoryReleasePredicate{
		releaseSet: ltsOnlyReleases,
	}
}

func (s *LTSOnlyRepositoryReleasePredicate) Apply(release *github.RepositoryRelease) bool {
	for _, r := range s.releaseSet {
		if r.Name == release.Name {
			return true
		}
	}

	return false
}
