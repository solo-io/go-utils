package debugutils

import (
	"context"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/installutils/kubeinstall"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/go-utils/testutils/clusterlock"
	"github.com/solo-io/go-utils/testutils/kube"
	"k8s.io/client-go/rest"
)

func TestDebugutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Debugutils Suite")
}

var (
	ns   string
	lock *clusterlock.TestClusterLocker


	restCfg     *rest.Config
	installer   kubeinstall.Installer
	manifests   helmchart.Manifests
	resources   kuberesource.UnstructuredResources
	ownerLabels map[string]string

	_ = SynchronizedBeforeSuite(func() []byte {
		var err error
		idPrefix := fmt.Sprintf("resource-collector-%s-%d-", os.Getenv("BUILD_ID"), config.GinkgoConfig.ParallelNode)
		lock, err = clusterlock.NewTestClusterLocker(kube.MustKubeClient(), clusterlock.Options{
			IdPrefix: idPrefix,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
		unique := "unique"
		randomLabel := testutils.RandString(8)
		ownerLabels = map[string]string{
			unique: randomLabel,
		}
		restCfg, err = kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		manifests, err = helmchart.RenderManifests(
			context.TODO(),
			"https://storage.googleapis.com/solo-public-helm/charts/gloo-0.13.33.tgz",
			"",
			"aaa",
			"gloo-system",
			"",
		)
		Expect(err).NotTo(HaveOccurred())
		cache := kubeinstall.NewCache()
		Expect(cache.Init(context.TODO(), restCfg)).NotTo(HaveOccurred())
		installer, err = kubeinstall.NewKubeInstaller(restCfg, cache, nil)
		Expect(err).NotTo(HaveOccurred())
		resources, err = manifests.ResourceList()
		Expect(err).NotTo(HaveOccurred())
		err = installer.ReconcileResources(context.TODO(), "gloo-system", resources, ownerLabels)
		Expect(err).NotTo(HaveOccurred())
		return nil
	}, func(data []byte) {})

	_ = SynchronizedAfterSuite(func() {}, func() {
		err := installer.PurgeResources(context.TODO(), ownerLabels)
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
	})
)
