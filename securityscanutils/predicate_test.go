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

func release(tag string) *github.RepositoryRelease {
	return &github.RepositoryRelease{TagName: github.String(tag)}
}

func releases(tags ...string) []*github.RepositoryRelease {
	out := make([]*github.RepositoryRelease, len(tags))
	for i, t := range tags {
		out[i] = release(t)
	}
	return out
}

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

		Context("with v-prefixed tags", func() {
			DescribeTable("single minor series",
				func(tag string, expected bool) {
					pred := securityscanutils.NewLatestPatchRepositoryReleasePredicate(
						releases("v1.0.1", "v1.0.2", "v1.0.3"))
					Expect(pred.Apply(release(tag))).To(Equal(expected))
				},
				Entry("latest patch is accepted", "v1.0.3", true),
				Entry("older patch is rejected", "v1.0.2", false),
				Entry("oldest patch is rejected", "v1.0.1", false),
				Entry("unknown version is rejected", "v3.0.0", false),
			)

			DescribeTable("multiple minor series",
				func(tag string, expected bool) {
					pred := securityscanutils.NewLatestPatchRepositoryReleasePredicate(
						releases("v1.0.0", "v1.0.1", "v1.1.0", "v1.1.1", "v2.0.0"))
					Expect(pred.Apply(release(tag))).To(Equal(expected))
				},
				Entry("latest patch of 1.0.x", "v1.0.1", true),
				Entry("older patch of 1.0.x", "v1.0.0", false),
				Entry("latest patch of 1.1.x", "v1.1.1", true),
				Entry("older patch of 1.1.x", "v1.1.0", false),
				Entry("latest patch of 2.0.x", "v2.0.0", true),
			)
		})

		Context("without v prefix", func() {
			DescribeTable("single minor series",
				func(tag string, expected bool) {
					pred := securityscanutils.NewLatestPatchRepositoryReleasePredicate(
						releases("2.0.0", "2.0.1", "2.0.2"))
					Expect(pred.Apply(release(tag))).To(Equal(expected))
				},
				Entry("latest patch is accepted", "2.0.2", true),
				Entry("older patch is rejected", "2.0.1", false),
				Entry("oldest patch is rejected", "2.0.0", false),
			)

			DescribeTable("multiple minor series",
				func(tag string, expected bool) {
					pred := securityscanutils.NewLatestPatchRepositoryReleasePredicate(
						releases("2.0.1", "2.1.2", "2.1.3"))
					Expect(pred.Apply(release(tag))).To(Equal(expected))
				},
				Entry("latest patch of 2.0.x", "2.0.1", true),
				Entry("older patch of 2.1.x", "2.1.2", false),
				Entry("latest patch of 2.1.x", "2.1.3", true),
			)

			DescribeTable("multiple major series",
				func(tag string, expected bool) {
					pred := securityscanutils.NewLatestPatchRepositoryReleasePredicate(
						releases("1.0.5", "1.1.3", "2.0.1", "2.1.0"))
					Expect(pred.Apply(release(tag))).To(Equal(expected))
				},
				Entry("latest patch of 1.0.x", "1.0.5", true),
				Entry("latest patch of 1.1.x", "1.1.3", true),
				Entry("latest patch of 2.0.x", "2.0.1", true),
				Entry("latest patch of 2.1.x", "2.1.0", true),
			)
		})

		Context("unsorted input", func() {
			It("sorts internally and picks correct latest patches", func() {
				// provide releases in random order
				pred := securityscanutils.NewLatestPatchRepositoryReleasePredicate(
					releases("1.0.0", "1.0.3", "1.0.1", "1.1.0", "1.0.2", "1.1.2", "1.1.1"))
				Expect(pred.Apply(release("1.0.3"))).To(BeTrue(), "1.0.3 should be latest of 1.0.x")
				Expect(pred.Apply(release("1.1.2"))).To(BeTrue(), "1.1.2 should be latest of 1.1.x")
				Expect(pred.Apply(release("1.0.2"))).To(BeFalse())
				Expect(pred.Apply(release("1.1.1"))).To(BeFalse())
				Expect(pred.Apply(release("1.0.0"))).To(BeFalse())
			})
		})

		Context("non-semver tags are skipped", func() {
			It("ignores invalid tags and still picks correct latest patches", func() {
				pred := securityscanutils.NewLatestPatchRepositoryReleasePredicate(
					releases("not-a-version", "2.0.1", "also-bad", "2.0.2"))
				Expect(pred.Apply(release("2.0.2"))).To(BeTrue())
				Expect(pred.Apply(release("2.0.1"))).To(BeFalse())
				Expect(pred.Apply(release("not-a-version"))).To(BeFalse())
			})
		})

		Context("empty release set", func() {
			It("rejects everything", func() {
				pred := securityscanutils.NewLatestPatchRepositoryReleasePredicate(releases())
				Expect(pred.Apply(release("1.0.0"))).To(BeFalse())
				Expect(pred.Apply(release("v1.0.0"))).To(BeFalse())
			})
		})

		Context("single release", func() {
			It("accepts the only release", func() {
				pred := securityscanutils.NewLatestPatchRepositoryReleasePredicate(
					releases("3.5.7"))
				Expect(pred.Apply(release("3.5.7"))).To(BeTrue())
			})
		})

		// Regression: before switching from versionutils.ParseVersion to semver.NewVersion,
		// non-v-prefixed tags silently failed parsing, producing an empty predicate map.
		// This is the bug visible in the log as:
		//   GithubIssueWriter configured with Predicate: &{releasesByTagName:map[]}
		// With an empty map the predicate rejects every release, so no GitHub issues
		// are created even when github-issue-minor output is configured.
		Context("non-v-prefixed tags must not produce an empty predicate (regression)", func() {
			It("accepts latest patches for strict-semver tags like those used by github-issue-minor repos", func() {
				// Simulate the exact flow from initializeRepoConfiguration (securityscan.go:201-203):
				//   issuePredicate = NewLatestPatchRepositoryReleasePredicate(releasesToScan)
				// using tags that lack the v prefix, as seen in production.
				pred := securityscanutils.NewLatestPatchRepositoryReleasePredicate(
					releases("2.0.0", "2.0.1", "2.1.0", "2.1.2", "2.1.3"))

				// The predicate must NOT be a no-op: latest patches must be accepted
				Expect(pred.Apply(release("2.0.1"))).To(BeTrue(), "latest patch of 2.0.x must be accepted")
				Expect(pred.Apply(release("2.1.3"))).To(BeTrue(), "latest patch of 2.1.x must be accepted")

				// Older patches must still be rejected
				Expect(pred.Apply(release("2.0.0"))).To(BeFalse())
				Expect(pred.Apply(release("2.1.2"))).To(BeFalse())
				Expect(pred.Apply(release("2.1.0"))).To(BeFalse())
			})
		})
	})

	Context("SortReleasesBySemver", func() {
		tags := func(rels []*github.RepositoryRelease) []string {
			out := make([]string, len(rels))
			for i, r := range rels {
				out[i] = r.GetTagName()
			}
			return out
		}

		It("sorts v-prefixed tags descending", func() {
			rels := releases("v1.0.0", "v2.0.0", "v1.1.0", "v1.0.1")
			githubutils.SortReleasesBySemver(rels)
			Expect(tags(rels)).To(Equal([]string{"v2.0.0", "v1.1.0", "v1.0.1", "v1.0.0"}))
		})

		It("sorts non-prefixed tags descending", func() {
			rels := releases("1.0.0", "2.0.0", "1.1.0", "1.0.1")
			githubutils.SortReleasesBySemver(rels)
			Expect(tags(rels)).To(Equal([]string{"2.0.0", "1.1.0", "1.0.1", "1.0.0"}))
		})

		It("sorts a realistic release set", func() {
			rels := releases("2.0.1", "2.1.3", "2.1.2", "2.0.0", "2.1.0", "2.1.1")
			githubutils.SortReleasesBySemver(rels)
			Expect(tags(rels)).To(Equal([]string{"2.1.3", "2.1.2", "2.1.1", "2.1.0", "2.0.1", "2.0.0"}))
		})

		It("handles pre-release tags", func() {
			rels := releases("2.0.0", "2.1.0-beta.1", "2.1.0-beta.2", "2.1.0")
			githubutils.SortReleasesBySemver(rels)
			Expect(tags(rels)).To(Equal([]string{"2.1.0", "2.1.0-beta.2", "2.1.0-beta.1", "2.0.0"}))
		})

		It("pushes non-semver tags to the end", func() {
			rels := releases("bad-tag", "2.0.0", "1.0.0")
			githubutils.SortReleasesBySemver(rels)
			Expect(tags(rels)).To(Equal([]string{"2.0.0", "1.0.0", "bad-tag"}))
		})

		It("handles empty list", func() {
			rels := releases()
			githubutils.SortReleasesBySemver(rels)
			Expect(rels).To(BeEmpty())
		})

		It("handles single element", func() {
			rels := releases("1.0.0")
			githubutils.SortReleasesBySemver(rels)
			Expect(tags(rels)).To(Equal([]string{"1.0.0"}))
		})
	})
})
