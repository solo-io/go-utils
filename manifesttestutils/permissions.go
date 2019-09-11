package manifesttestutils

import (
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
)

type ServiceAccountPermissions struct {
	serviceAccount map[string]*namespacePermissions
}

type namespacePermissions struct {
	namespace map[string]*apiGroupPermissions
}

type apiGroupPermissions struct {
	apiGroup map[string]*resourcePermissions
}

type resourcePermissions struct {
	resource map[string]*verbs
}

type verbs struct {
	values map[string]bool
}

func (p *ServiceAccountPermissions) AddExpectedPermission(serviceAccount, namespace string, apiGroups, resources, verbs []string) {
	if p.serviceAccount == nil {
		p.serviceAccount = make(map[string]*namespacePermissions)
	}
	if _, exists := p.serviceAccount[serviceAccount]; !exists {
		p.serviceAccount[serviceAccount] = &namespacePermissions{}
	}
	p.serviceAccount[serviceAccount].addExpectedPermission(namespace, apiGroups, resources, verbs)
}

func (p *namespacePermissions) addExpectedPermission(namespace string, apiGroups, resources, verbs []string) {
	if p.namespace == nil {
		p.namespace = make(map[string]*apiGroupPermissions)
	}
	if _, exists := p.namespace[namespace]; !exists {
		p.namespace[namespace] = &apiGroupPermissions{}
	}
	p.namespace[namespace].addExpectedPermission(apiGroups, resources, verbs)
}

func (p *apiGroupPermissions) addExpectedPermission(apiGroups, resources, verbs []string) {
	if p.apiGroup == nil {
		p.apiGroup = make(map[string]*resourcePermissions)
	}
	for _, g := range apiGroups {
		if _, exists := p.apiGroup[g]; !exists {
			p.apiGroup[g] = &resourcePermissions{}
		}
		p.apiGroup[g].addExpectedPermission(resources, verbs)
	}
}

func (p *resourcePermissions) addExpectedPermission(resources, verbsToAdd []string) {
	if p.resource == nil {
		p.resource = make(map[string]*verbs)
	}
	for _, r := range resources {
		if _, exists := p.resource[r]; !exists {
			p.resource[r] = &verbs{values: make(map[string]bool)}
		}
		for _, v := range verbsToAdd {
			p.resource[r].values[v] = true
		}
	}
}

func (t *testManifest) mustFindDeploymentsV1Beta1() []*v1beta1.Deployment {
	var deployments []*v1beta1.Deployment
	for _, resource := range t.resources {
		if resource.GetKind() == "Deployment" && resource.GetAPIVersion() == "extensions/v1beta1" {
			converted, err := kuberesource.ConvertUnstructured(resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(converted).To(BeAssignableToTypeOf(&v1beta1.Deployment{}))
			deployments = append(deployments, converted.(*v1beta1.Deployment))
		}
	}
	return deployments
}

func (t *testManifest) mustFindDeploymentsAppsV1() []*appsv1.Deployment {
	var deployments []*appsv1.Deployment
	for _, resource := range t.resources {
		if resource.GetKind() == "Deployment" && resource.GetAPIVersion() == "apps/v1" {
			converted, err := kuberesource.ConvertUnstructured(resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(converted).To(BeAssignableToTypeOf(&appsv1.Deployment{}))
			deployments = append(deployments, converted.(*appsv1.Deployment))
		}
	}
	return deployments
}

func (t *testManifest) mustFindServiceAccounts() []*corev1.ServiceAccount {
	var serviceAccounts []*corev1.ServiceAccount
	for _, resource := range t.resources {
		if resource.GetKind() == "ServiceAccount" {
			converted, err := kuberesource.ConvertUnstructured(resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(converted).To(BeAssignableToTypeOf(&corev1.ServiceAccount{}))
			serviceAccounts = append(serviceAccounts, converted.(*corev1.ServiceAccount))
		}
	}
	return serviceAccounts
}

// ApiGroup is "" for service accounts
func (t *testManifest) mustFindRoleBindings(subjectKind, subjectApiGroup, subjectNamespace, subjectName string) []*rbacv1.RoleBinding {
	var roleBindings []*rbacv1.RoleBinding
	for _, resource := range t.resources {
		if resource.GetKind() == "RoleBinding" {
			converted, err := kuberesource.ConvertUnstructured(resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(converted).To(BeAssignableToTypeOf(&rbacv1.RoleBinding{}))
			roleBinding := converted.(*rbacv1.RoleBinding)
			for _, s := range roleBinding.Subjects {
				if s.Kind == subjectKind && s.APIGroup == subjectApiGroup && s.Name == subjectName && s.Namespace == subjectNamespace {
					roleBindings = append(roleBindings, converted.(*rbacv1.RoleBinding))
				}
			}
		}
	}
	return roleBindings
}

// ApiGroup is "" for service accounts
func (t *testManifest) mustFindClusterRoleBindings(subjectKind, subjectApiGroup, subjectNamespace, subjectName string) []*rbacv1.ClusterRoleBinding {
	var clusterRoleBindings []*rbacv1.ClusterRoleBinding
	for _, resource := range t.resources {
		if resource.GetKind() == "ClusterRoleBinding" {
			converted, err := kuberesource.ConvertUnstructured(resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(converted).To(BeAssignableToTypeOf(&rbacv1.ClusterRoleBinding{}))
			roleBinding := converted.(*rbacv1.ClusterRoleBinding)
			for _, s := range roleBinding.Subjects {
				if s.Kind == subjectKind && s.APIGroup == subjectApiGroup && s.Name == subjectName && s.Namespace == subjectNamespace {
					clusterRoleBindings = append(clusterRoleBindings, converted.(*rbacv1.ClusterRoleBinding))
				}
			}
		}
	}
	return clusterRoleBindings
}
