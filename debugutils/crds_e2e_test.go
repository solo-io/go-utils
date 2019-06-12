package debugutils

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/installutils/kubeinstall"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/testutils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var _ = Describe("resource collector e2e", func() {
	var (
		restCfg       *rest.Config
		collector     *resourceCollector
		installer     kubeinstall.Installer
		manifests helmchart.Manifests
		resources     kuberesource.UnstructuredResources
		ownerLabels   map[string]string
	)

	BeforeSuite(func() {
		var err error
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
	})

	AfterSuite(func() {
		err := installer.PurgeResources(context.TODO(), ownerLabels)
		Expect(err).NotTo(HaveOccurred())
	})

	BeforeEach(func() {
		var err error
		collector, err = NewCrdCollector()
		Expect(err).NotTo(HaveOccurred())
	})

	Context("e2e", func() {
		It("can retrieve all gloo resources", func() {
			resources, err := collector.ResourcesFromManifest(manifests, v1.ListOptions{})
			Expect(err).NotTo(HaveOccurred())
			for _, resource := range resources {
				switch resource.GVK.Kind {
				case "ConfigMap":
					Expect(resource.Resources).To(HaveLen(1))
					Expect(resource.Resources[0].GetName()).To(Equal("gateway-proxy-envoy-config"))
				}
			}
		})
	})
})
