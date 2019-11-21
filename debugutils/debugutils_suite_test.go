package debugutils

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	"google.golang.org/api/iterator"
)

func TestDebugutils(t *testing.T) {
	T = t
	RegisterFailHandler(Fail)
	//RunSpecs(t, "Debugutils Suite")
}

var (
	T    *testing.T
	ns   string
	ctrl *gomock.Controller

	manifests             helmchart.Manifests
	unstructuredResources kuberesource.UnstructuredResources

	_ = SynchronizedBeforeSuite(func() []byte {
		var err error
		manifests, err = helmchart.RenderManifests(
			context.TODO(),
			"https://storage.googleapis.com/solo-public-helm/charts/gloo-0.13.33.tgz",
			"",
			"aaa",
			"gloo-system",
			"",
		)
		Expect(err).NotTo(HaveOccurred())
		unstructuredResources, err = manifests.ResourceList()
		Expect(err).NotTo(HaveOccurred())
		return nil
	}, func(data []byte) {})

	_ = SynchronizedAfterSuite(func() {}, func() {
		ctx := context.TODO()
		client, err := storage.NewClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		bucket := client.Bucket("go-utils-test")
		obj := os.ExpandEnv("$BUILD_ID")
		it := bucket.Objects(ctx, &storage.Query{
			Prefix: obj,
		})
		for {
			objAttrs, err := it.Next()
			if err != nil && err == iterator.Done {
				break
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(bucket.Object(objAttrs.Name).Delete(ctx)).NotTo(HaveOccurred())
		}
	})
)
