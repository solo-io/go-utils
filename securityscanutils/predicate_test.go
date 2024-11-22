package securityscanutils_test

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v32/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/securityscanutils"
)

var _ = Describe("Predicate", func() {
	Context("securityScanRepositoryReleasePredicate", func() {
		DescribeTable(
			"Returns true/false based on release properties",
			func(release *github.RepositoryRelease, enablePreRelease bool, expectedResult bool) {
				twoPlusConstraint, err := semver.NewConstraint(fmt.Sprintf(">= %s", "v2.0.0-0"))
				Expect(err).NotTo(HaveOccurred())

				releasePredicate := securityscanutils.NewSecurityScanRepositoryReleasePredicate(twoPlusConstraint, enablePreRelease)
				Expect(releasePredicate.Apply(release)).To(Equal(expectedResult))
			},
			Entry("release is draft", &github.RepositoryRelease{
				Draft: github.Bool(true),
			}, false, false),
			Entry("release tag does not respect semver", &github.RepositoryRelease{
				TagName: github.String("non-semver-tag-name"),
			}, false, false),
			Entry("release tag does not pass version constraint", &github.RepositoryRelease{
				TagName: github.String("v1.0.0"),
			}, false, false),
			Entry("release tag does pass version constraint", &github.RepositoryRelease{
				TagName: github.String("v2.0.1"),
			}, false, true),
			Entry("release tag has beta", &github.RepositoryRelease{
				TagName: github.String("v2.0.1-beta1"),
			}, false, true),
			Entry("release tag has rc", &github.RepositoryRelease{
				TagName: github.String("v2.0.1-rc2"),
			}, false, true),
			Entry("release is a pre-release", &github.RepositoryRelease{
				Prerelease: github.Bool(true),
			}, false, false),
			Entry("pre-release scan is enabled", &github.RepositoryRelease{
				TagName:    github.String("v2.0.1-alpha.0"),
				Prerelease: github.Bool(true),
			}, true, true),
		)
	})

	Context("latestPatchRepositoryReleasePredicate", func() {
		var releasePredicate githubutils.RepositoryReleasePredicate

		BeforeEach(func() {
			releaseSet := []*github.RepositoryRelease{
				{
					TagName: github.String("v1.0.1"),
				},
				{
					TagName: github.String("v1.0.2"),
				},
				{
					TagName: github.String("v1.0.3"),
				},
			}

			releasePredicate = securityscanutils.NewLatestPatchRepositoryReleasePredicate(releaseSet)
		})

		DescribeTable(
			"Returns true/false based on release properties",
			func(release *github.RepositoryRelease, expectedResult bool) {
				Expect(releasePredicate.Apply(release)).To(Equal(expectedResult))
			},
			Entry("release is not in release set", &github.RepositoryRelease{
				TagName: github.String("v3.0.0"), // not in the original release set
			}, false),
			Entry("release is not latest patch release", &github.RepositoryRelease{
				TagName: github.String("v1.0.1"),
			}, false),
			Entry("release is latest patch release", &github.RepositoryRelease{
				TagName: github.String("v1.0.3"),
			}, true),
		)
	})
})
