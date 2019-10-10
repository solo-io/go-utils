package githubutils_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/solo-io/go-utils/randutils"

	"github.com/google/go-github/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/versionutils"
)

var _ = Describe("github utils", func() {
	var (
		githubClient            *github.Client
		client                  githubutils.RepoClient
		ctx                     = context.Background()
		owner                   = "solo-io"
		repo                    = "testrepo"
		repoWithoutReleasesName = "testrepo-noreleases"
		sha                     = "9065a9a84e286ea7f067f4fc240944b0a4d4c82a"
		otherSha                = "ea649cd931820a6a59970b051d480094f9d61c4e"
		pr                      = 62
	)

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
		cc, err := client.CompareCommits(ctx, "6d389bc860e1cefdcbc99d43979e62104f13092f", "9065a9a84e286ea7f067f4fc240944b0a4d4c82a")
		Expect(err).To(BeNil())
		Expect(cc.Files).To(HaveLen(5))
	})

	It("can get sha for tag", func() {
		client = githubutils.NewRepoClient(githubClient, owner, repo)
		sha, err := client.GetShaForTag(ctx, "v0.1.16")
		Expect(err).To(BeNil())
		Expect(sha).To(Equal("04da4a385be3fde4797963cd4f3f76a185e56ba7"))
	})

	It("can get a commit", func() {
		client = githubutils.NewRepoClient(githubClient, owner, repo)
		commit, err := client.GetCommit(ctx, sha)
		Expect(err).To(BeNil())
		Expect(len(commit.Files)).To(Equal(3))
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

})
