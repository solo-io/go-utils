package manifesttestutils

import (
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
)

type RbacTester interface {
	ExpectServiceAccountPermissions(permissions *ServiceAccountPermissions)
}

type rbacTester struct {
	serviceAccountPermissions *ServiceAccountPermissions
}

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

func NewRbacTester(resources kuberesource.UnstructuredResources) RbacTester {
	manifest := testManifest{resources: resources}
	permissions := &ServiceAccountPermissions{}

	// get all deployments
	v1beta1Deployments := manifest.mustFindDeploymentsV1Beta1()
	appsv1Deployments := manifest.mustFindDeploymentsAppsV1()

	// get all service accounts referenced in deployments
	serviceAccounts := make([]*corev1.ServiceAccount, 0, len(v1beta1Deployments)+len(appsv1Deployments))
	for _, d := range v1beta1Deployments {
		if d.Spec.Template.Spec.ServiceAccountName == "" {
			continue
		}
		account := manifest.mustFindObject("ServiceAccount", d.Namespace, d.Spec.Template.Spec.ServiceAccountName)
		serviceAccounts = append(serviceAccounts, account.(*corev1.ServiceAccount))
	}
	for _, d := range appsv1Deployments {
		if d.Spec.Template.Spec.ServiceAccountName == "" {
			continue
		}
		account := manifest.mustFindObject("ServiceAccount", d.Namespace, d.Spec.Template.Spec.ServiceAccountName)
		serviceAccounts = append(serviceAccounts, account.(*corev1.ServiceAccount))
	}

	// get all roles
	for _, account := range serviceAccounts {
		roleBindings := manifest.mustFindRoleBindings("ServiceAccount", "", account.Namespace, account.Name)
		for _, rb := range roleBindings {
			obj := manifest.mustFindObject(rb.RoleRef.Kind, account.Namespace, rb.RoleRef.Name)
			Expect(obj).To(BeAssignableToTypeOf(&rbacv1.Role{}))
			role := obj.(*rbacv1.Role)
			for _, rule := range role.Rules {
				permissions.AddExpectedPermission(account.Namespace+"."+account.Name, account.Namespace, rule.APIGroups, rule.Resources, rule.Verbs)
			}
		}
	}

	// get all cluster roles
	for _, account := range serviceAccounts {
		clusterRoleBindings := manifest.mustFindClusterRoleBindings("ServiceAccount", "", account.Namespace, account.Name)
		for _, rb := range clusterRoleBindings {
			obj := manifest.mustFindObject(rb.RoleRef.Kind, "", rb.RoleRef.Name)
			Expect(obj).To(BeAssignableToTypeOf(&rbacv1.ClusterRole{}))
			clusterRole := obj.(*rbacv1.ClusterRole)
			for _, rule := range clusterRole.Rules {
				permissions.AddExpectedPermission(account.Namespace+"."+account.Name, corev1.NamespaceAll, rule.APIGroups, rule.Resources, rule.Verbs)
			}
		}
	}

	return &rbacTester{
		serviceAccountPermissions: permissions,
	}
}

func (r rbacTester) ExpectServiceAccountPermissions(permissions *ServiceAccountPermissions) {
	Expect(r.serviceAccountPermissions).To(BeEquivalentTo(permissions))

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
