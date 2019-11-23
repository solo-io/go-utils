package manifesttestutils

import (
	"fmt"
	"io/ioutil"

	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"github.com/ghodss/yaml"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"

	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	yaml2json "k8s.io/apimachinery/pkg/util/yaml"
)

type TestManifest interface {
	// Deprecated
	ExpectDeployment(deployment *v1beta1.Deployment) *v1beta1.Deployment
	ExpectDeploymentAppsV1(deployment *appsv1.Deployment)
	ExpectServiceAccount(serviceAccount *corev1.ServiceAccount)
	ExpectClusterRole(clusterRole *rbacv1.ClusterRole)
	ExpectClusterRoleBinding(clusterRoleBinding *rbacv1.ClusterRoleBinding)
	ExpectRole(role *rbacv1.Role)
	ExpectRoleBinding(roleBinding *rbacv1.RoleBinding)
	ExpectConfigMap(configMap *corev1.ConfigMap)
	ExpectConfigMapWithYamlData(configMap *corev1.ConfigMap)
	ExpectSecret(secret *corev1.Secret)
	ExpectService(service *corev1.Service)
	ExpectNamespace(namespace *corev1.Namespace)
	ExpectCrd(crd *extv1beta1.CustomResourceDefinition)
	ExpectCustomResource(gvk, namespace, name string) *unstructured.Unstructured
	NumResources() int

	Expect(kind, namespace, name string) Assertion
	ExpectUnstructured(kind, namespace, name string) Assertion

	ExpectPermissions(permissions *ServiceAccountPermissions)

	// run this callback on all the resources contained in this TestManifest
	ExpectAll(callback func(*unstructured.Unstructured))

	// construct a new set of resources to make assertions against by collecting
	// all the resources for which `selector` returns true
	SelectResources(selector func(*unstructured.Unstructured) bool) TestManifest
}

type testManifest struct {
	resources kuberesource.UnstructuredResources
}

func NewTestManifest(relativePathToManifest string) TestManifest {
	return &testManifest{
		resources: mustGetResources(relativePathToManifest),
	}
}

func NewTestManifestWithResources(resources kuberesource.UnstructuredResources) TestManifest {
	return &testManifest{
		resources: resources,
	}
}

func (t *testManifest) NumResources() int {
	return len(t.resources)
}

func (t *testManifest) ExpectDeployment(deployment *v1beta1.Deployment) *v1beta1.Deployment {
	obj := t.mustFindObject(deployment.Kind, deployment.Namespace, deployment.Name)
	Expect(obj).To(BeAssignableToTypeOf(&v1beta1.Deployment{}))
	actual := obj.(*v1beta1.Deployment)
	Expect(actual).To(BeEquivalentTo(deployment))
	return actual
}

func (t *testManifest) ExpectDeploymentAppsV1(deployment *appsv1.Deployment) {
	obj := t.mustFindObject(deployment.Kind, deployment.Namespace, deployment.Name)
	Expect(obj).To(BeAssignableToTypeOf(&appsv1.Deployment{}))
	actual := obj.(*appsv1.Deployment)
	Expect(actual).To(BeEquivalentTo(deployment))
}

func (t *testManifest) ExpectServiceAccount(serviceAccount *corev1.ServiceAccount) {
	obj := t.mustFindObject(serviceAccount.Kind, serviceAccount.Namespace, serviceAccount.Name)
	Expect(obj).To(BeAssignableToTypeOf(&corev1.ServiceAccount{}))
	actual := obj.(*corev1.ServiceAccount)
	Expect(actual).To(BeEquivalentTo(serviceAccount))
}

func (t *testManifest) ExpectClusterRole(clusterRole *rbacv1.ClusterRole) {
	obj := t.mustFindObject(clusterRole.Kind, clusterRole.Namespace, clusterRole.Name)
	Expect(obj).To(BeAssignableToTypeOf(&rbacv1.ClusterRole{}))
	actual := obj.(*rbacv1.ClusterRole)
	Expect(actual).To(BeEquivalentTo(clusterRole))
}

func (t *testManifest) ExpectClusterRoleBinding(clusterRoleBinding *rbacv1.ClusterRoleBinding) {
	obj := t.mustFindObject(clusterRoleBinding.Kind, clusterRoleBinding.Namespace, clusterRoleBinding.Name)
	Expect(obj).To(BeAssignableToTypeOf(&rbacv1.ClusterRoleBinding{}))
	actual := obj.(*rbacv1.ClusterRoleBinding)
	Expect(actual).To(BeEquivalentTo(clusterRoleBinding))
}

func (t *testManifest) ExpectRole(role *rbacv1.Role) {
	obj := t.mustFindObject(role.Kind, role.Namespace, role.Name)
	Expect(obj).To(BeAssignableToTypeOf(&rbacv1.Role{}))
	actual := obj.(*rbacv1.Role)
	Expect(actual).To(BeEquivalentTo(role))
}

func (t *testManifest) ExpectRoleBinding(roleBinding *rbacv1.RoleBinding) {
	obj := t.mustFindObject(roleBinding.Kind, roleBinding.Namespace, roleBinding.Name)
	Expect(obj).To(BeAssignableToTypeOf(&rbacv1.RoleBinding{}))
	actual := obj.(*rbacv1.RoleBinding)
	Expect(actual).To(BeEquivalentTo(roleBinding))
}

func (t *testManifest) ExpectConfigMap(configMap *corev1.ConfigMap) {
	obj := t.mustFindObject(configMap.Kind, configMap.Namespace, configMap.Name)
	Expect(obj).To(BeAssignableToTypeOf(&corev1.ConfigMap{}))
	actual := obj.(*corev1.ConfigMap)
	Expect(actual).To(BeEquivalentTo(configMap))
}

func (t *testManifest) ExpectConfigMapWithYamlData(configMap *corev1.ConfigMap) {
	obj := t.mustFindObject(configMap.Kind, configMap.Namespace, configMap.Name)
	Expect(obj).To(BeAssignableToTypeOf(&corev1.ConfigMap{}))
	actual := obj.(*corev1.ConfigMap)
	for k, v := range actual.Data {
		actual.Data[k] = MustCanonicalizeYaml(v)
	}
	for k, v := range configMap.Data {
		configMap.Data[k] = MustCanonicalizeYaml(v)
	}
	Expect(actual).To(BeEquivalentTo(configMap))
}

func (t *testManifest) ExpectSecret(secret *corev1.Secret) {
	obj := t.mustFindObject(secret.Kind, secret.Namespace, secret.Name)
	Expect(obj).To(BeAssignableToTypeOf(&corev1.Secret{}))
	actual := obj.(*corev1.Secret)
	Expect(actual).To(BeEquivalentTo(secret))
}

func (t *testManifest) ExpectService(service *corev1.Service) {
	obj := t.mustFindObject(service.Kind, service.Namespace, service.Name)
	Expect(obj).To(BeAssignableToTypeOf(&corev1.Service{}))
	actual := obj.(*corev1.Service)
	Expect(actual).To(BeEquivalentTo(service))
}

func (t *testManifest) ExpectNamespace(namespace *corev1.Namespace) {
	obj := t.mustFindObject(namespace.Kind, "", namespace.Name)
	Expect(obj).To(BeAssignableToTypeOf(&corev1.Namespace{}))
	actual := obj.(*corev1.Namespace)
	Expect(actual).To(BeEquivalentTo(namespace))
}

func (t *testManifest) ExpectCrd(crd *extv1beta1.CustomResourceDefinition) {
	obj := t.mustFindObject(crd.Kind, "", crd.Name)
	Expect(obj).To(BeAssignableToTypeOf(&extv1beta1.CustomResourceDefinition{}))
	actual := obj.(*extv1beta1.CustomResourceDefinition)
	Expect(actual).To(BeEquivalentTo(crd))
}

func (t *testManifest) ExpectCustomResource(kind, namespace, name string) *unstructured.Unstructured {
	found := false
	for _, resource := range t.resources {
		if resource.GetKind() == kind && resource.GetNamespace() == namespace && resource.GetName() == name {
			found = true
			return resource
		}
	}
	Expect(found).To(BeTrue())
	return nil
}

func (t *testManifest) Expect(kind, namespace, name string) GomegaAssertion {
	return Expect(t.findObject(kind, namespace, name))
}

func (t *testManifest) ExpectUnstructured(kind, namespace, name string) GomegaAssertion {
	return Expect(t.findUnstructured(kind, namespace, name))
}

func (t *testManifest) ExpectPermissions(permissions *ServiceAccountPermissions) {
	manifestPermissions := &ServiceAccountPermissions{}

	// get all deployments
	v1beta1Deployments := t.mustFindDeploymentsV1Beta1()
	appsv1Deployments := t.mustFindDeploymentsAppsV1()

	// get all service accounts referenced in deployments
	serviceAccounts := make([]*corev1.ServiceAccount, 0, len(v1beta1Deployments)+len(appsv1Deployments))
	for _, d := range v1beta1Deployments {
		if d.Spec.Template.Spec.ServiceAccountName == "" {
			continue
		}
		account := t.mustFindObject("ServiceAccount", d.Namespace, d.Spec.Template.Spec.ServiceAccountName)
		serviceAccounts = append(serviceAccounts, account.(*corev1.ServiceAccount))
	}
	for _, d := range appsv1Deployments {
		if d.Spec.Template.Spec.ServiceAccountName == "" {
			continue
		}
		account := t.mustFindObject("ServiceAccount", d.Namespace, d.Spec.Template.Spec.ServiceAccountName)
		serviceAccounts = append(serviceAccounts, account.(*corev1.ServiceAccount))
	}

	// get all roles
	for _, account := range serviceAccounts {
		roleBindings := t.mustFindRoleBindings("ServiceAccount", "", account.Namespace, account.Name)
		for _, rb := range roleBindings {
			obj := t.mustFindObject(rb.RoleRef.Kind, account.Namespace, rb.RoleRef.Name)
			Expect(obj).To(BeAssignableToTypeOf(&rbacv1.Role{}))
			role := obj.(*rbacv1.Role)
			for _, rule := range role.Rules {
				manifestPermissions.AddExpectedPermission(account.Namespace+"."+account.Name, account.Namespace, rule.APIGroups, rule.Resources, rule.Verbs)
			}
		}
	}

	// get all cluster roles
	for _, account := range serviceAccounts {
		clusterRoleBindings := t.mustFindClusterRoleBindings("ServiceAccount", "", account.Namespace, account.Name)
		for _, rb := range clusterRoleBindings {
			obj := t.mustFindObject(rb.RoleRef.Kind, "", rb.RoleRef.Name)
			Expect(obj).To(BeAssignableToTypeOf(&rbacv1.ClusterRole{}))
			clusterRole := obj.(*rbacv1.ClusterRole)
			for _, rule := range clusterRole.Rules {
				manifestPermissions.AddExpectedPermission(account.Namespace+"."+account.Name, corev1.NamespaceAll, rule.APIGroups, rule.Resources, rule.Verbs)
			}
		}
	}

	// Convert permission structs to YAML for:
	// 1) correct assertions
	// 2) readable failures
	ownYaml, err := yaml.Marshal(manifestPermissions)
	Expect(err).NotTo(HaveOccurred())
	expectedYaml, err := yaml.Marshal(permissions)
	Expect(err).NotTo(HaveOccurred())

	Expect(string(ownYaml)).To(BeEquivalentTo(string(expectedYaml)))
}

func (t *testManifest) SelectResources(selector func(*unstructured.Unstructured) bool) TestManifest {
	selectedResources := &testManifest{}

	for _, resource := range t.resources {
		if selector(resource) {
			selectedResources.resources = append(selectedResources.resources, resource)
		}
	}

	return selectedResources
}

func (t *testManifest) ExpectAll(callback func(*unstructured.Unstructured)) {
	for _, resource := range t.resources {
		callback(resource)
	}
}

func (t *testManifest) findUnstructured(kind, namespace, name string) *unstructured.Unstructured {
	for _, resource := range t.resources {
		if resource.GetKind() == kind && resource.GetNamespace() == namespace && resource.GetName() == name {
			return resource
		}
	}
	return nil
}

func (t *testManifest) findObject(kind, namespace, name string) runtime.Object {
	if unst := t.findUnstructured(kind, namespace, name); unst != nil {
		converted, err := kuberesource.ConvertUnstructured(unst)
		Expect(err).NotTo(HaveOccurred())
		return converted
	}
	return nil
}

func (t *testManifest) mustFindObject(kind, namespace, name string) runtime.Object {
	obj := t.findObject(kind, namespace, name)
	if obj == nil {
		Fail(fmt.Sprintf("can't find object %s %s %s", kind, namespace, name))
	}
	return obj
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
		// yaml2json errors are very terse and expect the source yaml to be available when interpreting the failure
		// message, so print it on failure
		infoForError(err, objectYaml)
		Expect(err).To(BeNil())

		uncastObj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, jsn)
		infoForError(err, objectYaml)
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

func infoForError(err error, str string) {
	if err != nil {
		fmt.Printf("error is: %v\nrelated info:\n%v\n", err, str)
	}
}
