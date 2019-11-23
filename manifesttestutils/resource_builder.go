package manifesttestutils

import (
	"fmt"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceBuilder struct {
	Namespace          string
	Name               string
	Args               []string
	Labels             map[string]string
	Annotations        map[string]string
	Rules              []rbacv1.PolicyRule
	Data               map[string]string
	Subjects           []rbacv1.Subject
	RoleRef            rbacv1.RoleRef
	Containers         []ContainerSpec
	Service            ServiceSpec
	SecretType         v1.SecretType
	ServiceAccountName string
}

type ContainerSpec struct {
	Image      string
	PullPolicy v1.PullPolicy
	Name       string
	Args       []string
	EnvVars    []v1.EnvVar
}

type ServiceSpec struct {
	Ports []PortSpec
}

type PortSpec struct {
	Name string
	Port int
}

// Deprecated
func (b *ResourceBuilder) GetDeployment() *v1beta1.Deployment {
	return &v1beta1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.Name,
			Namespace: b.Namespace,
			Labels:    b.Labels,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: getReplicas(),
			Selector: &metav1.LabelSelector{
				MatchLabels: b.Labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      b.Labels,
					Annotations: b.Annotations,
				},
				Spec: v1.PodSpec{
					ServiceAccountName: b.ServiceAccountName,
					Containers:         b.getContainers(),
				},
			},
		},
	}
}

func (b *ResourceBuilder) GetDeploymentAppsv1() *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.Name,
			Namespace: b.Namespace,
			Labels:    b.Labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: getReplicas(),
			Selector: &metav1.LabelSelector{
				MatchLabels: b.Labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      b.Labels,
					Annotations: b.Annotations,
				},
				Spec: v1.PodSpec{
					Containers: b.getContainers(),
				},
			},
		},
	}
}

func (b *ResourceBuilder) GetServiceAccount() *v1.ServiceAccount {
	return &v1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.Name,
			Namespace:   b.Namespace,
			Labels:      b.Labels,
			Annotations: b.Annotations,
		},
	}
}

func (b *ResourceBuilder) GetNamespace() *v1.Namespace {
	annotations := b.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["helm.sh/hook"] = "pre-install"
	return &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.Name,
			Labels:      b.Labels,
			Annotations: annotations,
		},
	}
}

func (b *ResourceBuilder) GetClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.Name,
			Namespace:   b.Namespace,
			Labels:      b.Labels,
			Annotations: b.Annotations,
		},
		Rules: b.Rules,
	}
}

func (b *ResourceBuilder) GetClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.Name,
			Namespace:   b.Namespace,
			Labels:      b.Labels,
			Annotations: b.Annotations,
		},
		Subjects: b.Subjects,
		RoleRef:  b.RoleRef,
	}
}

func (b *ResourceBuilder) GetRole() *rbacv1.Role {
	return &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.Name,
			Namespace:   b.Namespace,
			Labels:      b.Labels,
			Annotations: b.Annotations,
		},
		Rules: b.Rules,
	}
}

func (b *ResourceBuilder) GetRoleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.Name,
			Namespace:   b.Namespace,
			Labels:      b.Labels,
			Annotations: b.Annotations,
		},
		Subjects: b.Subjects,
		RoleRef:  b.RoleRef,
	}
}

func (b *ResourceBuilder) GetConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		Data: b.Data,
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.Name,
			Namespace:   b.Namespace,
			Labels:      b.Labels,
			Annotations: b.Annotations,
		},
	}
}

func (b *ResourceBuilder) GetSecret() *v1.Secret {
	byteMap := make(map[string][]byte)
	for k, v := range b.Data {
		byteMap[k] = []byte(v)
	}
	secretType := b.SecretType
	if secretType == "" {
		secretType = v1.SecretTypeOpaque
	}
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.Name,
			Namespace:   b.Namespace,
			Labels:      b.Labels,
			Annotations: b.Annotations,
		},
		Data: byteMap,
		Type: secretType,
	}
}

func (b *ResourceBuilder) GetService() *v1.Service {
	var ports []v1.ServicePort
	for _, spec := range b.Service.Ports {
		ports = append(ports, b.getPort(spec))
	}
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.Name,
			Namespace:   b.Namespace,
			Labels:      b.Labels,
			Annotations: b.Annotations,
		},
		Spec: v1.ServiceSpec{
			Selector: b.Labels,
			Type:     v1.ServiceTypeNodePort,
			Ports:    ports,
		},
	}
}

func (b *ResourceBuilder) getPort(spec PortSpec) v1.ServicePort {
	if spec.Name == "" {
		spec.Name = "static"
	}
	return v1.ServicePort{
		Protocol: v1.ProtocolTCP,
		Name:     spec.Name,
		Port:     int32(spec.Port),
	}
}

func (b *ResourceBuilder) getContainers() []v1.Container {
	var containers []v1.Container
	for _, spec := range b.Containers {
		containers = append(containers, b.getContainer(spec))
	}
	return containers
}

func (b *ResourceBuilder) getContainer(spec ContainerSpec) v1.Container {
	if spec.Name == "" {
		spec.Name = b.Name
	}
	if spec.PullPolicy == "" {
		spec.PullPolicy = v1.PullIfNotPresent
	}
	return v1.Container{
		Image:           spec.Image,
		ImagePullPolicy: spec.PullPolicy,
		Name:            spec.Name,
		Args:            spec.Args,
		Env:             spec.EnvVars,
	}
}

func GetPodNamespaceEnvVar() v1.EnvVar {
	return v1.EnvVar{
		Name: "POD_NAMESPACE",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "metadata.namespace",
			},
		},
	}
}

func getReplicas() *int32 {
	replicas := int32(1)
	return &replicas
}

func getReadOnlyVerbs() []string {
	return []string{"get", "list", "watch"}
}

func (b *ResourceBuilder) GetClusterRoleRef() rbacv1.RoleRef {
	return rbacv1.RoleRef{
		Kind:     "ClusterRole",
		Name:     b.Name,
		APIGroup: "rbac.authorization.k8s.io",
	}
}

func (b *ResourceBuilder) GetServiceAccountSubject() rbacv1.Subject {
	return rbacv1.Subject{
		Name:      b.Name,
		Kind:      "ServiceAccount",
		Namespace: b.Namespace,
	}
}

func GetCrdRule() rbacv1.PolicyRule {
	return rbacv1.PolicyRule{
		APIGroups: []string{"apiextensions.k8s.io"},
		Resources: []string{"customresourcedefinitions"},
		Verbs:     []string{"get", "create"},
	}
}

func GetCoreRule() rbacv1.PolicyRule {
	return rbacv1.PolicyRule{
		APIGroups: []string{""},
		Resources: []string{"configmaps", "pods", "services", "secrets", "endpoints", "namespaces"},
		Verbs:     getReadOnlyVerbs(),
	}
}

func GetIstioRule() rbacv1.PolicyRule {
	return rbacv1.PolicyRule{
		APIGroups: []string{"authentication.istio.io"},
		Resources: []string{"meshpolicies"},
		Verbs:     getReadOnlyVerbs(),
	}
}

func GetDefaultServiceAccountSubject(namespace string) rbacv1.Subject {
	return GetServiceAccountSubject(namespace, "default")
}

func GetServiceAccountSubject(namespace, name string) rbacv1.Subject {
	subjectBuilder := ResourceBuilder{
		Name:      name,
		Namespace: namespace,
	}
	return subjectBuilder.GetServiceAccountSubject()
}

func GetClusterRoleRef(name string) rbacv1.RoleRef {
	refBuilder := ResourceBuilder{
		Name: name,
	}
	return refBuilder.GetClusterRoleRef()
}

func GetClusterAdminRoleRef() rbacv1.RoleRef {
	return GetClusterRoleRef("cluster-admin")
}

func GetAppLabels(appName, category string) map[string]string {
	return map[string]string{
		"app":   appName,
		appName: category,
	}
}

func GetContainerSpec(registry, name, tag string, envVars ...v1.EnvVar) ContainerSpec {
	return ContainerSpec{
		Name:    name,
		Image:   fmt.Sprintf("%s:%s", filepath.Join(registry, name), tag),
		EnvVars: envVars,
	}
}

func GetQuayContainerSpec(image, tag string, envVars ...v1.EnvVar) ContainerSpec {
	return GetContainerSpec("quay.io/solo-io", image, tag, envVars...)
}
