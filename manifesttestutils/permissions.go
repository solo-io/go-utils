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
	ServiceAccount map[string]*NamespacePermissions
}

type NamespacePermissions struct {
	Namespace map[string]*ApiGroupPermissions
}

type ApiGroupPermissions struct {
	ApiGroup map[string]*ResourcePermissions
}

type ResourcePermissions struct {
	Resource map[string]*Verbs
}

type Verbs struct {
	Values map[string]bool
}

func (p *ServiceAccountPermissions) AddExpectedPermission(serviceAccount, namespace string, apiGroups, resources, verbs []string) {
	if p.ServiceAccount == nil {
		p.ServiceAccount = make(map[string]*NamespacePermissions)
	}
	if _, exists := p.ServiceAccount[serviceAccount]; !exists {
		p.ServiceAccount[serviceAccount] = &NamespacePermissions{}
	}
	p.ServiceAccount[serviceAccount].addExpectedPermission(namespace, apiGroups, resources, verbs)
}

func (p *NamespacePermissions) addExpectedPermission(namespace string, apiGroups, resources, verbs []string) {
	if p.Namespace == nil {
		p.Namespace = make(map[string]*ApiGroupPermissions)
	}
	if _, exists := p.Namespace[namespace]; !exists {
		p.Namespace[namespace] = &ApiGroupPermissions{}
	}
	p.Namespace[namespace].addExpectedPermission(apiGroups, resources, verbs)
}

func (p *ApiGroupPermissions) addExpectedPermission(apiGroups, resources, verbs []string) {
	if p.ApiGroup == nil {
		p.ApiGroup = make(map[string]*ResourcePermissions)
	}
	for _, g := range apiGroups {
		if _, exists := p.ApiGroup[g]; !exists {
			p.ApiGroup[g] = &ResourcePermissions{}
		}
		p.ApiGroup[g].addExpectedPermission(resources, verbs)
	}
}

func (p *ResourcePermissions) addExpectedPermission(resources, verbsToAdd []string) {
	if p.Resource == nil {
		p.Resource = make(map[string]*Verbs)
	}
	for _, r := range resources {
		if _, exists := p.Resource[r]; !exists {
			p.Resource[r] = &Verbs{Values: make(map[string]bool)}
		}
		for _, v := range verbsToAdd {
			p.Resource[r].Values[v] = true
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
