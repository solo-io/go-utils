package vfsutils_test

import (
	"context"
	"os"

	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/vfsutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("mounted repo utils", func() {

	const (
		owner = "solo-io"
	)

	var (
		ctx             = context.Background()
		mountedRepo     vfsutils.MountedRepo
		repo            = "testrepo"
		sha             = "9065a9a84e286ea7f067f4fc240944b0a4d4c82a"
		file            = "tmp.txt"
		expectedContent = "another"
		path            = "namespace"
	)

	if os.Getenv("HAS_CLOUDBUILD_GITHUB_TOKEN") == "" {
		repo = "unik"
		sha = "767fb7285ea9c893efcced90a612c4e253ef8e4b"
		file = "README.md"
		expectedContent = "UniK"
		path = "containers/utils/vsphere-client/src/main/java/com/emc/unik"
	}

	BeforeEach(func() {
		client, err := githubutils.GetClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		mountedRepo = vfsutils.NewLazilyMountedRepo(client, owner, repo, sha)
	})

	It("can get contents", func() {
		contents, err := mountedRepo.GetFileContents(ctx, file)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(ContainSubstring(expectedContent))
	})

	It("can list files", func() {
		files, err := mountedRepo.ListFiles(ctx, path)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(files)).To(Equal(1))
	})

})
