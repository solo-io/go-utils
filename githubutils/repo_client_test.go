package githubutils_test

import (
	"context"

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
})
