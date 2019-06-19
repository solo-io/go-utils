package test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/debugutils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("resource collector e2e", func() {
	var (
		collector debugutils.ResourceCollector
	)

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
			var err error
			collector, err = debugutils.DefaultResourceCollector()
			Expect(err).NotTo(HaveOccurred())
		})
		It("can retrieve all gloo resources", func() {
			unstructured, err := manifests.ResourceList()
			Expect(err).NotTo(HaveOccurred())
			collectedResources, err := collector.RetrieveResources(unstructured, "", v1.ListOptions{})
			Expect(err).NotTo(HaveOccurred())
			for _, resource := range collectedResources {
				switch resource.GVK.Kind {
				case "ConfigMap":
					Expect(resource.Resources).To(HaveLen(1))
					Expect(resource.Resources[0].GetName()).To(Equal("gateway-proxy-envoy-config"))
				case "Pod":
					Expect(resource.Resources).To(HaveLen(4))
					var deploymentNames []string
					for _, v := range unstructuredResources {
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
