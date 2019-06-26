package kubeutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/kubeutils"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("WaitCrd", func() {
	var (
		api     apiexts.Interface
		crdName = "testing"
	)
	BeforeEach(func() {
		cfg, err := GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		api, err = apiexts.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		crd, err := api.ApiextensionsV1beta1().CustomResourceDefinitions().Create(&v1beta1.CustomResourceDefinition{
			ObjectMeta: v1.ObjectMeta{Name: "somethings.test.solo.io"},
			Spec: v1beta1.CustomResourceDefinitionSpec{
				Group: "test.solo.io",
				Names: v1beta1.CustomResourceDefinitionNames{
					Plural:     "somethings",
					Kind:       "Something",
					ShortNames: []string{"st"},
				},
				Version: "v1",
			},
		})
		Expect(err).NotTo(HaveOccurred())
		crdName = crd.Name
	})
	AfterEach(func() {
		api.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(crdName, nil)
	})
	It("waits successfully for a crd to become established", func() {
		err := WaitForCrdActive(api, crdName)
		Expect(err).NotTo(HaveOccurred())
	})
})
