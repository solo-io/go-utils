package githubutils

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/google/go-github/v32/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/versionutils"
)

var _ = Describe("github utils", func() {
	var (
		client                  *github.Client
		ctx                     = context.Background()
		owner                   = "solo-io"
		reponame                = "testrepo"
		repoWithoutReleasesName = "testrepo-noreleases"
		version                 = "v0.0.16"
		ref                     = "v0.0.17"
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

	It("can get 'latest release version' for repo with no prior releases", func() {
		version, err := FindLatestReleaseTagIncudingPrerelease(ctx, client, owner, repoWithoutReleasesName)
		Expect(err).NotTo(HaveOccurred())
		Expect(version).To(Equal(versionutils.SemverNilVersionValue))
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

	It("can download and store archive from git", func() {
		file, dir := mustSetupTempFiles()
		defer os.Remove(dir)
		defer os.Remove(file.Name())
		err := DownloadRepoArchive(ctx, client, file, owner, reponame, ref)
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()
		info, err := file.Stat()
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Size()).To(BeNumerically(">", 0))
	})

})

func mustSetupTempFiles() (file *os.File, dir string) {
	tmpf, err := ioutil.TempFile("", "tar-file-")
	Expect(err).NotTo(HaveOccurred())
	tmpd, err := ioutil.TempDir("", "tar-dir-")
	Expect(err).NotTo(HaveOccurred())
	return tmpf, tmpd
}
