package test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/debugutils"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("pod unit tests", func() {
	Context("Label Pod Finder", func() {
		var (
			podFinder *debugutils.LabelPodFinder
			clientset *fake.Clientset
		)

		BeforeEach(func() {
			clientset = fake.NewSimpleClientset(podsAsObjects(GeneratePodList())...)
			podFinder = debugutils.NewLabelPodFinder(clientset)
		})

		It("can handle full use case", func() {
			resources, err := manifests.ResourceList()
			Expect(err).NotTo(HaveOccurred())
			list, err := podFinder.GetPods(resources)
			Expect(err).NotTo(HaveOccurred())
			Expect(list).To(HaveLen(4))
			for _, v := range list {
				Expect(v.Items).NotTo(HaveLen(0))
			}
		})

		It("can work with an individual pod", func() {
			var unstructuredPod unstructured.Unstructured
			err := yaml.Unmarshal([]byte(GlooPodYaml), &unstructuredPod)
			Expect(err).NotTo(HaveOccurred())
			list, err := podFinder.GetPods(kuberesource.UnstructuredResources{&unstructuredPod})
			Expect(err).NotTo(HaveOccurred())
			Expect(list).To(HaveLen(1))
			Expect(list[0].Items).To(HaveLen(1))
			pod := list[0].Items[0]
			Expect(pod.GetName()).To(ContainSubstring("gloo"))
		})
	})
})

func podsAsObjects(list *corev1.PodList) []runtime.Object {
	result := make([]runtime.Object, len(list.Items))
	for i, v := range list.Items {
		v := v
		result[i] = &v
	}
	return result
}
