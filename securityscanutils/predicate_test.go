package securityscanutils_test

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v32/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/securityscanutils"
)

var _ = Describe("Predicate", func() {

	Context("securityScanRepositoryReleasePredicate", func() {

		var (
			releasePredicate githubutils.RepositoryReleasePredicate
		)

		BeforeEach(func() {
			twoPlusConstraint, err := semver.NewConstraint(fmt.Sprintf(">= %s", "v2.0.0"))
			Expect(err).NotTo(HaveOccurred())

			releasePredicate = securityscanutils.NewSecurityScanRepositoryReleasePredicate(twoPlusConstraint)
		})

		DescribeTable(
			"Returns true/false based on release properties",
			func(release *github.RepositoryRelease, expectedResult bool) {
				Expect(releasePredicate.Apply(release)).To(Equal(expectedResult))
			},
			Entry("release is draft", &github.RepositoryRelease{
				Draft: github.Bool(true),
			}, false),
			Entry("release is pre-release", &github.RepositoryRelease{
				Prerelease: github.Bool(true),
			}, false),
			Entry("release tag does not respect semver", &github.RepositoryRelease{
				TagName: github.String("non-semver-tag-name"),
			}, false),
			Entry("release tag does not pass version constraint", &github.RepositoryRelease{
				TagName: github.String("v1.0.0"),
			}, false),
			Entry("release tag does pass version constraint", &github.RepositoryRelease{
				TagName: github.String("v2.0.1"),
			}, true),
		)

	})

	Context("latestPatchRepositoryReleasePredicate", func() {

		var (
			releasePredicate githubutils.RepositoryReleasePredicate
		)

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
