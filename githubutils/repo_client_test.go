package githubutils_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/solo-io/go-utils/randutils"

	"github.com/google/go-github/v32/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/versionutils"
)

var _ = Describe("repo client utils", func() {
	var (
		githubClient            *github.Client
		client                  githubutils.RepoClient
		ctx                     = context.Background()
		owner                   = "solo-io"
		repo                    = "testrepo"
		repoWithoutReleasesName = "testrepo-noreleases"
		sha                     = "9065a9a84e286ea7f067f4fc240944b0a4d4c82a"
		commitsInSha            = 3
		otherSha                = "ea649cd931820a6a59970b051d480094f9d61c4e"
		pr                      = 62
		commit1                 = "6d389bc860e1cefdcbc99d43979e62104f13092f"
		commit2                 = "9065a9a84e286ea7f067f4fc240944b0a4d4c82a"
		tagWithSha              = "v0.1.16"
		shaForTag               = "04da4a385be3fde4797963cd4f3f76a185e56ba7"
	)

	if os.Getenv("HAS_CLOUDBUILD_GITHUB_TOKEN") == "" {
		repo = "reporting-client"
		repoWithoutReleasesName = "unik-hub"
		sha = "af5d207720ee6b548704b06bfa6631f9a2897294"
		commitsInSha = 2
		commit1 = "7ef898bc3df32db0e1ed2dee70a838c955a7b422"
		commit2 = "f47eacc21bd62e6bc8bb8954af0dc1817079af0d"
		tagWithSha = "v0.1.2"
		shaForTag = "a1c75ffaa40ea2b89368bfc338dc3f6f990b6df2"
	}

	BeforeEach(func() {
		c, err := githubutils.GetClient(ctx)
		Expect(err).To(BeNil())
		githubClient = c
	})

	It("can get latest release version", func() {
		client = githubutils.NewRepoClient(githubClient, owner, repo)
		version, err := client.FindLatestReleaseTagIncudingPrerelease(ctx)
		Expect(err).NotTo(HaveOccurred())
		_, err = versionutils.ParseVersion(version)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can get 'latest release version' for repo with no prior releases", func() {
		client = githubutils.NewRepoClient(githubClient, owner, repoWithoutReleasesName)
		version, err := client.FindLatestReleaseTagIncudingPrerelease(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(version).To(Equal(versionutils.SemverNilVersionValue))
		_, err = versionutils.ParseVersion(version)
		Expect(err).NotTo(HaveOccurred())
	})

	It("finds a directory that exists", func() {
		client = githubutils.NewRepoClient(githubClient, owner, repo)
		exists, err := client.DirectoryExists(ctx, sha, "changelog")
		Expect(err).To(BeNil())
		Expect(exists).To(BeTrue())
	})

	It("doesn't find a directory that doesn't exist", func() {
		client = githubutils.NewRepoClient(githubClient, owner, repo)
		exists, err := client.DirectoryExists(ctx, sha, "doesnt-exist")
		Expect(err).To(BeNil())
		Expect(exists).To(BeFalse())
	})

	It("can do a commit comparison", func() {
		client = githubutils.NewRepoClient(githubClient, owner, repo)
		cc, err := client.CompareCommits(ctx, commit1, commit2)
		Expect(err).To(BeNil())
		Expect(cc.Files).To(HaveLen(5))
	})

	It("can get sha for tag", func() {
		client = githubutils.NewRepoClient(githubClient, owner, repo)
		sha, err := client.GetShaForTag(ctx, tagWithSha)
		Expect(err).To(BeNil())
		Expect(sha).To(Equal(shaForTag))
	})

	It("can get a commit", func() {
		client = githubutils.NewRepoClient(githubClient, owner, repo)
		commit, err := client.GetCommit(ctx, sha)
		Expect(err).To(BeNil())
		Expect(len(commit.Files)).To(Equal(commitsInSha))
	})

	expectStatus := func(actual, expected *github.RepoStatus) {
		Expect(actual.State).To(BeEquivalentTo(expected.State))
		Expect(actual.Description).To(BeEquivalentTo(expected.Description))
		Expect(actual.Context).To(BeEquivalentTo(expected.Context))
	}

	testManageStatus := func(client githubutils.RepoClient, status *github.RepoStatus, commitSha string) {
		stored, err := client.CreateStatus(ctx, commitSha, status)
		Expect(err).To(BeNil())
		expectStatus(stored, status)
		loaded, err := client.FindStatus(ctx, status.GetContext(), commitSha)
		Expect(err).To(BeNil())
		expectStatus(loaded, status)
	}

	It("can manage status", func() {
		client = githubutils.NewRepoClient(githubClient, owner, repo)
		// randomizing this would create a (slim) potential race
		// not randomizing makes it less easy to validate that the create worked, but we'll assume github api responses are accurate
		status := &github.RepoStatus{
			State:       github.String(githubutils.STATUS_SUCCESS),
			Context:     github.String("test"),
			Description: github.String("test"), // longer than 140 characters will be truncated
		}
		testManageStatus(client, status, sha)
	})

	It("can manage status even when it exceeds 140 character", func() {
		client = githubutils.NewRepoClient(githubClient, owner, repo)
		// randomizing this would create a (slim) potential race
		// not randomizing makes it less easy to validate that the create worked, but we'll assume github api responses are accurate
		status := &github.RepoStatus{
			State:       github.String(githubutils.STATUS_SUCCESS),
			Context:     github.String("test"),
			Description: github.String(strings.Repeat("test", 40)), // longer than 140 characters will be truncated
		}
		testManageStatus(client, status, otherSha) // don't share sha with other test to avoid race
	})

	It("can create and delete comments", func() {
		client = githubutils.NewRepoClient(githubClient, owner, repo)
		body := fmt.Sprintf("test-%s", randutils.RandString(4))
		comment := &github.IssueComment{
			Body: github.String(body),
		}
		stored, err := client.CreateComment(ctx, pr, comment)
		Expect(err).To(BeNil())
		err = client.DeleteComment(ctx, stored.GetID())
		Expect(err).To(BeNil())
	})

	Context("Can properly find the most recent tag before an SHA", func() {
		// This test is pretty slow, these are cumbersome calls but wanted to make sure it works, including paging when necessary
		// As more releases are cut, the no of API requests can grow due to pagination - this can lead to CI failing
		// `403 API rate limit of 5000 still exceeded until 2023-11-28 17:49:31 +0000 UTC, not making remote request. [rate reset in 7m52s]`
		// To prevent this failure, this test needs to be periodically updated to test against more recent releases
		// To update, find the most recent release for solo-io/gloo matching each criterion and use that as the expected release,
		// checkout the branch it is on, or the prior release, choose either the release commit, if the test case is an exact
		// match, or a commit 2-4 after the chosen release if the test case is a "before" case, and use that as the input SHA
		BeforeEach(func() {
			client = githubutils.NewRepoClient(githubClient, owner, "gloo")
		})

		It("properly finds the most recent GA release tag matching an SHA", func() {
			tag, err := client.FindLatestTagIncludingPrereleaseBeforeSha(ctx, "e658203d0a0b7b479cbb59cfc43832699d25fb1c")
			Expect(err).To(BeNil())
			Expect(tag).To(Equal("v1.17.8"))
		})

		It("properly finds the most recent beta release tag before an SHA", func() {
			tag, err := client.FindLatestTagIncludingPrereleaseBeforeSha(ctx, "33cc7ee95c7319d33c36fb7d449a933dca95d211")
			Expect(err).To(BeNil())
			Expect(tag).To(Equal("v1.18.0-beta21"))
		})

		It("properly finds the most recent pre-release release tag before an SHA", func() {
			tag, err := client.FindLatestTagIncludingPrereleaseBeforeSha(ctx, "3e00d8140f91fe0111955bb46fbc29df8008bf47")
			Expect(err).To(BeNil())
			Expect(tag).To(Equal("v1.17.0-beta18"))
		})

		It("properly finds the most recent RC release tag before an SHA", func() {
			tag, err := client.FindLatestTagIncludingPrereleaseBeforeSha(ctx, "3067c264aa2025a31c7de82b8878b388d5bd0c4b")
			Expect(err).To(BeNil())
			Expect(tag).To(Equal("v1.17.0-rc12"))
		})

		// for this case, use a release that is not found on the first page of the API endpoint results here:
		// https://api.github.com/repos/solo-io/gloo/releases
		It("properly finds the most recent release tag before an SHA with pagination", func() {
			tag, err := client.FindLatestTagIncludingPrereleaseBeforeSha(ctx, "51cc97a355236c7f725fbf43fbee276a0208d12d")
			Expect(err).To(BeNil())
			Expect(tag).To(Equal("v1.18.0-beta7"))
		})
	})
})
