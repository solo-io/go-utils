package githubutils

import (
	"context"

	"github.com/google/go-github/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/versionutils"
)

var _ = Describe("github utils", func() {
	var (
		client   *github.Client
		ctx      = context.Background()
		owner    = "solo-io"
		reponame = "testrepo"
		version  = "v0.0.16"
		ref      = "v0.0.17"
	)

	var _ = BeforeEach(func() {
		var err error
		client, err = GetClient(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can get latest release version", func() {
		version, err := FindLatestReleaseTag(ctx, client, owner, reponame)
		Expect(err).NotTo(HaveOccurred())
		_, err = versionutils.ParseVersion(version)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can get all changelog files", func() {
		_, err := GetFilesForChangelogVersion(ctx, client, owner, reponame, ref, version)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can download bytes for single file", func() {
		files, err := GetFilesForChangelogVersion(ctx, client, owner, reponame, ref, version)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(files)).To(BeNumerically(">", 0))
		byt, err := GetRawGitFile(ctx, client, files[0], owner, reponame, ref)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(byt)).To(BeNumerically(">", 0))
	})

})
