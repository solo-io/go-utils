package builders

import (
	"context"
	"path"

	"cloud.google.com/go/storage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("gcloud unit tests", func() {

	var (
		ctx = context.Background()
	)

	It("path.Split has expected functionality", func() {
		uri := "gs://one/two/three-22.tgz"

		dir, file := path.Split(uri)
		Expect(file).To(Equal("three-22.tgz"))
		Expect(dir).To(Equal("gs://one/two/"))
	})

	Context("builders", func() {
		var builderCtx *mockBuilderContext
		BeforeEach(func() {
			var err error
			builderCtx, err = DefaultMockBuilderContext(ctx)
			Expect(err).NotTo(HaveOccurred())
		})
		Context("storage source", func() {
			var (
				sb *StorageBuilder
			)

			BeforeEach(func() {
				var err error
				client, err := storage.NewClient(ctx)
				Expect(err).NotTo(HaveOccurred())

				sb = &StorageBuilder{
					client: client,
				}
			})
			It("can init build with sha", func() {
				_, err := sb.InitBuildWithSha(ctx, builderCtx)
				Expect(err).NotTo(HaveOccurred())
			})
			It("can init build with tag", func() {
				_, err := sb.InitBuildWithTag(ctx, builderCtx)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("repo source", func() {

		})
	})
})
