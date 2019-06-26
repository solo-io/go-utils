package vfsutils_test

import (
	"context"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/vfsutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("mounted repo utils", func() {

	const (
		owner = "solo-io"
		repo  = "testrepo"
		sha   = "9065a9a84e286ea7f067f4fc240944b0a4d4c82a"
	)

	var (
		ctx         = context.Background()
		mountedRepo vfsutils.MountedRepo
	)

	BeforeEach(func() {
		client, err := githubutils.GetClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		mountedRepo = vfsutils.NewLazilyMountedRepo(client, owner, repo, sha)
	})

	It("can get contents", func() {
		contents, err := mountedRepo.GetFileContents(ctx, "tmp.txt")
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(ContainSubstring("another"))
	})

	It("can list files", func() {
		files, err := mountedRepo.ListFiles(ctx, "namespace")
		Expect(err).NotTo(HaveOccurred())
		Expect(len(files)).To(Equal(1))
	})

})
