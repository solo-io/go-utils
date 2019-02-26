package kubeinstallutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/solo-io/go-utils/kubeinstallutils"
	"github.com/solo-io/go-utils/kubeutils"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// source: https://raw.githubusercontent.com/linkerd/linkerd2/master/cli/install/template.go
const linkerdCrdYaml = `### Service Profile CRD ###
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: serviceprofiles.linkerd.io
spec:
  group: linkerd.io
  version: v1alpha1
  scope: Namespaced
  names:
    plural: serviceprofiles
    singular: serviceprofile
    kind: ServiceProfile 
    shortNames:
    - sp
`

var _ = Describe("Crd", func() {
	var (
		testCrds []*v1beta1.CustomResourceDefinition
		apiExts  apiexts.Interface
	)
	BeforeEach(func() {
		cfg, err := kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		apiExts, err = apiexts.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		testCrds, err = kubeinstallutils.CrdsFromManifest(linkerdCrdYaml)
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		var crdsToDelete []string
		for _, crd := range testCrds {
			crdsToDelete = append(crdsToDelete, crd.Name)
		}
		kubeinstallutils.DeleteCrds(apiExts, crdsToDelete...)
	})
	It("creates crds", func() {
		err := kubeinstallutils.CreateCrds(apiExts, testCrds...)
		Expect(err).NotTo(HaveOccurred())
		crdList, err := apiExts.ApiextensionsV1beta1().CustomResourceDefinitions().List(v1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		for _, testCrd := range testCrds {
			var found bool
			for _, actual := range crdList.Items {
				if testCrd.Name == actual.Name {
					found = true
					// set by apiserver
					testCrd.Spec.Names.ListKind = testCrd.Spec.Names.Kind + "List"
					// Kubernetes now creates their CRDs with slightly more detailed info
					// TODO: decide if we should create CRDs with that extra info
					Expect(testCrd.Spec).To(MatchFields(IgnoreExtras, Fields{
						"Group": Equal(actual.Spec.Group),
						"Names": Equal(actual.Spec.Names),
						"Scope": Equal(actual.Spec.Scope),
					}))
					break
				}
			}
			Expect(found).To(BeTrue())
		}
	})
})
