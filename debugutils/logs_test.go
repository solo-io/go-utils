package debugutils

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("logs", func() {
	FContext("request builder", func() {
		var (
			requestBuilder *LogRequestBuilder

			deployedPods = []*LogsRequest{
				{
					podMeta: metav1.ObjectMeta{
						Name: "gateway",
					},
					containerName: "gateway",
				},
				{
					podMeta: metav1.ObjectMeta{
						Name: "gateway-proxy",
					},
					containerName: "gateway-proxy",
				},
				{
					podMeta: metav1.ObjectMeta{
						Name: "gloo",
					},
					containerName: "gloo",
				},
				{
					podMeta: metav1.ObjectMeta{
						Name: "discovery",
					},
					containerName: "discovery",
				},
			}

		)
		BeforeEach(func() {
			var err error
			requestBuilder, err = NewLogRequestBuilder()
			Expect(err).NotTo(HaveOccurred())
		})
		It("can properly build the requests from the gloo manifest", func() {
			requests, err := requestBuilder.LogsFromManifest(manifests)
			Expect(err).NotTo(HaveOccurred())
			Expect(requests).To(HaveLen(4))
			for _, deployedPod := range deployedPods {
				found := false
				for _, request := range requests {
					if request.containerName == deployedPod.containerName &&
						strings.HasPrefix(request.podMeta.Name, deployedPod.podMeta.Name) {
						found = true
						continue
					}
				}
				Expect(found).To(BeTrue())
			}
		})
	})
})