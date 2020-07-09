package debugutils

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("storage client tests", func() {

	var (
		storageObjects []*StorageObject
	)
	BeforeEach(func() {
		storageObjects = []*StorageObject{
			{
				Resource: bytes.NewBufferString("first"),
				Name:     "first",
			},
			{
				Resource: bytes.NewBufferString("second"),
				Name:     "second",
			},
			{
				Resource: bytes.NewBufferString("third"),
				Name:     "third",
			},
		}
	})

	Context("file client", func() {
		var (
			client *FileStorageClient
			fs     afero.Fs
			tmpd   string
		)

		BeforeEach(func() {
			var err error
			fs = afero.NewOsFs()
			client = NewFileStorageClient(fs)
			tmpd, err = afero.TempDir(fs, "", "")
			Expect(err).NotTo(HaveOccurred())

		})

		AfterEach(func() {
			fs.RemoveAll(tmpd)
		})

		It("can store a single file", func() {
			Expect(client.Save(tmpd, storageObjects[0])).NotTo(HaveOccurred())
			fileByt, err := afero.ReadFile(fs, filepath.Join(tmpd, storageObjects[0].Name))
			Expect(err).NotTo(HaveOccurred())
			Expect(fileByt).To(Equal([]byte(storageObjects[0].Name)))
		})

		It("can store multiple files", func() {
			Expect(client.Save(tmpd, storageObjects...)).NotTo(HaveOccurred())
			for _, v := range storageObjects {
				fileByt, err := afero.ReadFile(fs, filepath.Join(tmpd, v.Name))
				Expect(err).NotTo(HaveOccurred())
				Expect(fileByt).To(Equal([]byte(v.Name)))
			}
		})

		It("can store no files", func() {
			Expect(client.Save(tmpd)).NotTo(HaveOccurred())
		})

	})

	Context("gcs client", func() {
		const (
			bucketName = "go-utils-test"
		)
		var (
			client *GcsStorageClient
			ctx    context.Context
			bucket *storage.BucketHandle

			buildIdFile = func(resourceName string) string {
				buildId := os.Getenv("BUILD_ID")
				if buildId == "" {
					Skip("the gcs storage client tests require build id")
				}
				return fmt.Sprintf("%s/%s", buildId, resourceName)
			}
		)

		BeforeEach(func() {
			var err error
			ctx = context.Background()
			client, err = DefaultGcsStorageClient(ctx)
			Expect(err).NotTo(HaveOccurred())
			bucket = client.client.Bucket(bucketName)
		})

		AfterEach(func() {
			// obj := os.ExpandEnv("$BUILD_ID")
			// it := bucket.Objects(ctx, &storage.Query{
			// 	Prefix: obj,
			// })
			// for {
			// 	objAttrs, err := it.Next()
			// 	if err != nil && err == iterator.Done {
			// 		break
			// 	}
			// 	Expect(err).NotTo(HaveOccurred())
			// 	Expect(bucket.Object(objAttrs.Name).Delete(ctx)).NotTo(HaveOccurred())
			// }
		})

		It("can store a single file", func() {
			storageObjects[0].Name = buildIdFile(storageObjects[0].Name)
			Expect(client.Save(bucketName, storageObjects[0])).NotTo(HaveOccurred())
			obj := bucket.Object(storageObjects[0].Name)
			reader, err := obj.NewReader(ctx)
			defer reader.Close()
			Expect(err).NotTo(HaveOccurred())
			byt, err := ioutil.ReadAll(reader)
			Expect(err).NotTo(HaveOccurred())
			Expect(byt).To(ContainSubstring(filepath.Base(storageObjects[0].Name)))
		})

		It("can store multiple files", func() {
			for _, v := range storageObjects {
				v.Name = buildIdFile(v.Name)
			}
			Expect(client.Save(bucketName, storageObjects...)).NotTo(HaveOccurred())
			for _, v := range storageObjects {
				obj := bucket.Object(v.Name)
				reader, err := obj.NewReader(ctx)
				defer reader.Close()
				Expect(err).NotTo(HaveOccurred())
				byt, err := ioutil.ReadAll(reader)
				Expect(err).NotTo(HaveOccurred())
				Expect(byt).To(ContainSubstring(filepath.Base(v.Name)))
			}
		})

		It("can store no files", func() {
			Expect(client.Save(bucketName)).NotTo(HaveOccurred())
		})
	})
})
