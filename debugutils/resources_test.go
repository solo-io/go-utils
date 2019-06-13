package debugutils

import (
	"context"
	"strings"

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
		restCfg     *rest.Config
		collector   *resourceCollector
		installer   kubeinstall.Installer
		manifests   helmchart.Manifests
		resources   kuberesource.UnstructuredResources
		ownerLabels map[string]string
	)


	SynchronizedBeforeSuite(func() []byte {
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
		return nil
	}, func(data []byte) {})

	SynchronizedAfterSuite(func() {}, func() {
		err := installer.PurgeResources(context.TODO(), ownerLabels)
		Expect(err).NotTo(HaveOccurred())
	})

	var (
		containsPrefixToString = func(s string, prefixes []string) bool {
			for _, prefix := range prefixes {
				if strings.HasPrefix(s, prefix) {
					return true
				}
			}
			return false
		}
	)

	Context("e2e", func() {
		BeforeEach(func() {
			err := installer.ReconcileResources(context.TODO(), "gloo-system", resources, ownerLabels)
			Expect(err).NotTo(HaveOccurred())
			collector, err = NewResourceCollector()
			Expect(err).NotTo(HaveOccurred())
		})
		It("can retrieve all gloo resources", func() {
			collectedResources, err := collector.ResourcesFromManifest(manifests, v1.ListOptions{})
			Expect(err).NotTo(HaveOccurred())
			for _, resource := range collectedResources {
				switch resource.GVK.Kind {
				case "ConfigMap":
					Expect(resource.Resources).To(HaveLen(1))
					Expect(resource.Resources[0].GetName()).To(Equal("gateway-proxy-envoy-config"))
				case "Pod":
					Expect(resource.Resources).To(HaveLen(4))
					var deploymentNames []string
					for _, v := range resources {
						if v.GetKind() == "Deployment" {
							deploymentNames = append(deploymentNames, v.GetName())
						}
					}
					var podNames []string
					for _, v := range resource.Resources {
						podNames = append(podNames, v.GetName())
					}
					for _, v := range podNames {
						Expect(containsPrefixToString(v, deploymentNames)).To(BeTrue())
					}
				}

			}
		})
	})
})
