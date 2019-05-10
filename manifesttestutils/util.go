package manifesttestutils

import (
	"io/ioutil"

	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"github.com/ghodss/yaml"

	"k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"

	"regexp"

	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	yaml2json "k8s.io/apimachinery/pkg/util/yaml"
)

type TestManifest interface {
	ExpectDeployment(deployment *v1beta1.Deployment)
	ExpectServiceAccount(serviceAccount *corev1.ServiceAccount)
	ExpectClusterRole(clusterRole *rbacv1.ClusterRole)
	ExpectClusterRoleBinding(clusterRoleBinding *rbacv1.ClusterRoleBinding)
	ExpectConfigMap(configMap *corev1.ConfigMap)
	ExpectConfigMapWithYamlData(configMap *corev1.ConfigMap)
	ExpectService(service *corev1.Service)
	ExpectNamespace(namespace *corev1.Namespace)
	ExpectCrd(crd *extv1beta1.CustomResourceDefinition)
	NumResources() int
}

type testManifest struct {
	resources kuberesource.UnstructuredResources
}

func NewTestManifest(relativePathToManifest string) TestManifest {
	return &testManifest{
		resources: mustGetResources(relativePathToManifest),
	}
}

func (t *testManifest) NumResources() int {
	return len(t.resources)
}

func (t *testManifest) ExpectDeployment(deployment *v1beta1.Deployment) {
	obj := t.mustFindObject(deployment.Kind, deployment.Namespace, deployment.Name)
	actual, ok := obj.(*v1beta1.Deployment)
	Expect(ok).To(BeTrue())
	Expect(actual).To(BeEquivalentTo(deployment))
}

func (t *testManifest) ExpectServiceAccount(serviceAccount *corev1.ServiceAccount) {
	obj := t.mustFindObject(serviceAccount.Kind, serviceAccount.Namespace, serviceAccount.Name)
	actual, ok := obj.(*corev1.ServiceAccount)
	Expect(ok).To(BeTrue())
	Expect(actual).To(BeEquivalentTo(serviceAccount))
}

func (t *testManifest) ExpectClusterRole(clusterRole *rbacv1.ClusterRole) {
	obj := t.mustFindObject(clusterRole.Kind, clusterRole.Namespace, clusterRole.Name)
	actual, ok := obj.(*rbacv1.ClusterRole)
	Expect(ok).To(BeTrue())
	Expect(actual).To(BeEquivalentTo(clusterRole))
}

func (t *testManifest) ExpectClusterRoleBinding(clusterRoleBinding *rbacv1.ClusterRoleBinding) {
	obj := t.mustFindObject(clusterRoleBinding.Kind, clusterRoleBinding.Namespace, clusterRoleBinding.Name)
	actual, ok := obj.(*rbacv1.ClusterRoleBinding)
	Expect(ok).To(BeTrue())
	Expect(actual).To(BeEquivalentTo(clusterRoleBinding))
}

func (t *testManifest) ExpectConfigMap(configMap *corev1.ConfigMap) {
	obj := t.mustFindObject(configMap.Kind, configMap.Namespace, configMap.Name)
	actual, ok := obj.(*corev1.ConfigMap)
	Expect(ok).To(BeTrue())
	Expect(actual).To(BeEquivalentTo(configMap))
}

func (t *testManifest) ExpectConfigMapWithYamlData(configMap *corev1.ConfigMap) {
	obj := t.mustFindObject(configMap.Kind, configMap.Namespace, configMap.Name)
	actual, ok := obj.(*corev1.ConfigMap)
	Expect(ok).To(BeTrue())
	for k, v := range actual.Data {
		actual.Data[k] = MustCanonicalizeYaml(v)
	}
	for k, v := range configMap.Data {
		configMap.Data[k] = MustCanonicalizeYaml(v)
	}
	Expect(actual).To(BeEquivalentTo(configMap))
}

func (t *testManifest) ExpectService(service *corev1.Service) {
	obj := t.mustFindObject(service.Kind, service.Namespace, service.Name)
	actual, ok := obj.(*corev1.Service)
	Expect(ok).To(BeTrue())
	Expect(actual).To(BeEquivalentTo(service))
}

func (t *testManifest) ExpectNamespace(namespace *corev1.Namespace) {
	obj := t.mustFindObject(namespace.Kind, "", namespace.Name)
	actual, ok := obj.(*corev1.Namespace)
	Expect(ok).To(BeTrue())
	Expect(actual).To(BeEquivalentTo(namespace))
}

func (t *testManifest) ExpectCrd(crd *extv1beta1.CustomResourceDefinition) {
	obj := t.mustFindObject(crd.Kind, "", crd.Name)
	actual, ok := obj.(*extv1beta1.CustomResourceDefinition)
	Expect(ok).To(BeTrue())
	Expect(actual).To(BeEquivalentTo(crd))
}

func (t *testManifest) mustFindObject(kind, namespace, name string) runtime.Object {
	for _, resource := range t.resources {
		if resource.GetKind() == kind && resource.GetNamespace() == namespace && resource.GetName() == name {
			converted, err := kuberesource.ConvertUnstructured(resource)
			Expect(err).NotTo(HaveOccurred())
			return converted
		}
	}
	Expect(false).To(BeTrue())
	return nil
}

func mustReadManifest(relativePathToManifest string) string {
	bytes, err := ioutil.ReadFile(relativePathToManifest)
	Expect(err).NotTo(HaveOccurred())
	return string(bytes)
}

var (
	yamlSeparator = regexp.MustCompile("\n---")
)

func mustGetResources(relativePathToManifest string) kuberesource.UnstructuredResources {
	manifest := mustReadManifest(relativePathToManifest)
	snippets := yamlSeparator.Split(manifest, -1)

	var resources kuberesource.UnstructuredResources
	for _, objectYaml := range snippets {
		if helmchart.IsEmptyManifest(objectYaml) {
			continue
		}
		jsn, err := yaml2json.ToJSON([]byte(objectYaml))
		Expect(err).To(BeNil())

		uncastObj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, jsn)
		Expect(err).To(BeNil())
		if resourceList, ok := uncastObj.(*unstructured.UnstructuredList); ok {
			for _, item := range resourceList.Items {
				resources = append(resources, &item)
			}
			continue
		}
		resources = append(resources, uncastObj.(*unstructured.Unstructured))
	}
	return resources
}

func MustCanonicalizeYaml(input string) string {
	jsn, err := yaml.YAMLToJSON([]byte(input))
	Expect(err).To(BeNil())
	yml, err := yaml.JSONToYAML(jsn)
	Expect(err).To(BeNil())
	return string(yml)
}
