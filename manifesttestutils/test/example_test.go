package test

import (
	. "github.com/solo-io/go-utils/manifesttestutils"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helm Test", func() {

	appName := "sm-marketplace"
	namespace := "sm-marketplace"
	version := "dev"

	It("has the right number of resources", func() {
		Expect(testManifest.NumResources()).To(Equal(14))
	})

	Describe("namespace and crds", func() {
		labels := map[string]string{
			"app": appName,
		}

		It("has a namespace", func() {
			rb := ResourceBuilder{
				Name:   namespace,
				Labels: labels,
			}
			testManifest.ExpectNamespace(rb.GetNamespace())
		})

		crd := v1beta1.CustomResourceDefinition{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CustomResourceDefinition",
				APIVersion: "apiextensions.k8s.io/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "applicationstates.marketplace.solo.io",
				Annotations: map[string]string{
					"helm.sh/hook": "crd-install",
				},
				Labels: labels,
			},
			Spec: v1beta1.CustomResourceDefinitionSpec{
				Group: "marketplace.solo.io",
				Names: v1beta1.CustomResourceDefinitionNames{
					Kind:       "ApplicationState",
					ListKind:   "ApplicationStateList",
					Plural:     "applicationstates",
					ShortNames: []string{"appstate"},
				},
				Scope:   "Namespaced",
				Version: "v1",
			},
		}

		It("has crds", func() {
			testManifest.ExpectCrd(&crd)
		})
	})

	Describe("rbac", func() {

		labels := GetAppLabels(appName, "rbac")

		It("has the default role binding", func() {
			rb := ResourceBuilder{
				Name:     "smm-default-role-binding",
				Labels:   labels,
				Subjects: []rbacv1.Subject{GetDefaultServiceAccountSubject(namespace)},
				RoleRef:  GetClusterAdminRoleRef(),
			}
			testManifest.ExpectClusterRoleBinding(rb.GetClusterRoleBinding())
		})
	})

	Describe("apiserver", func() {

		name := "apiserver"
		image := "smm-" + name
		labels := GetAppLabels(appName, name)
		annotations := map[string]string{
			"demo": "annotation",
		}

		It("has a deployment", func() {
			grpcPort := v1.EnvVar{
				Name:  "GRPC_PORT",
				Value: "10101",
			}
			apiserverContainer := GetQuayContainerSpec(image, version, GetPodNamespaceEnvVar(), grpcPort)
			envoyContainer := GetQuayContainerSpec("smm-envoy", version)
			uiContainer := GetQuayContainerSpec("smm-ui", version)
			rb := ResourceBuilder{
				Name:        image,
				Namespace:   namespace,
				Annotations: annotations,
				Labels:      labels,
				Containers:  []ContainerSpec{apiserverContainer, uiContainer, envoyContainer},
			}
			testManifest.ExpectDeployment(rb.GetDeployment())
		})

		It("has a service", func() {
			port := PortSpec{
				Name: "static",
				Port: 8080,
			}
			rb := ResourceBuilder{
				Name:        image,
				Namespace:   namespace,
				Labels:      labels,
				Annotations: annotations,
				Service: ServiceSpec{
					Ports: []PortSpec{port},
				},
			}
			testManifest.ExpectService(rb.GetService())
		})

		It("has a config with yaml data", func() {
			configString := "registries:\n- name: default\n  github:\n    org: solo-io\n    repo: service-mesh-hub\n    ref: master\n    directory: extensions/v1\n"
			configString = MustCanonicalizeYaml(configString)
			data := map[string]string{
				"config.yaml": configString,
			}
			rb := ResourceBuilder{
				Name:      name + "-config",
				Namespace: namespace,
				Labels:    labels,
				Data:      data,
			}
			testManifest.ExpectConfigMapWithYamlData(rb.GetConfigMap())
		})
	})

	Describe("operator", func() {

		name := "operator"
		image := "smm-" + name
		labels := GetAppLabels(appName, name)

		It("has a deployment", func() {
			operatorContainer := GetQuayContainerSpec(image, version, GetPodNamespaceEnvVar())
			operatorContainer.PullPolicy = v1.PullAlways
			rb := ResourceBuilder{
				Name:       image,
				Namespace:  namespace,
				Labels:     labels,
				Containers: []ContainerSpec{operatorContainer},
			}
			testManifest.ExpectDeployment(rb.GetDeployment())
		})

		It("has a apps/v1 deployment", func() {
			operatorContainer := GetQuayContainerSpec(image, version, GetPodNamespaceEnvVar())
			operatorContainer.PullPolicy = v1.PullAlways
			rb := ResourceBuilder{
				Name:       image + "-apps-v1",
				Namespace:  namespace,
				Labels:     labels,
				Containers: []ContainerSpec{operatorContainer},
			}
			testManifest.ExpectDeploymentAppsV1(rb.GetDeploymentAppsv1())
		})

		It("has a config map", func() {
			data := map[string]string{
				"refreshRate": "1s",
			}
			rb := ResourceBuilder{
				Name:      name + "-config",
				Namespace: namespace,
				Labels:    labels,
				Data:      data,
			}
			testManifest.ExpectConfigMap(rb.GetConfigMap())
		})
	})

	Describe("mesh discovery", func() {

		name := "mesh-discovery"
		labels := GetAppLabels(appName, name)

		It("has a service account", func() {
			rb := ResourceBuilder{
				Name:      name,
				Namespace: namespace,
				Labels:    labels,
			}
			testManifest.ExpectServiceAccount(rb.GetServiceAccount())
		})

		It("has a deployment", func() {
			container := GetQuayContainerSpec(name, "0.3.13", GetPodNamespaceEnvVar())
			container.Args = []string{"--disable-config"}
			rb := ResourceBuilder{
				Labels:             labels,
				Name:               name,
				Namespace:          namespace,
				Containers:         []ContainerSpec{container},
				ServiceAccountName: "mesh-discovery",
			}
			testManifest.ExpectDeployment(rb.GetDeployment())
		})

		It("has a cluster role", func() {
			rules := []rbacv1.PolicyRule{
				GetCoreRule(),
				GetIstioRule(),
				GetCrdRule(),
			}
			rb := ResourceBuilder{
				Name:   name,
				Labels: labels,
				Rules:  rules,
			}
			testManifest.ExpectClusterRole(rb.GetClusterRole())
		})

		It("has a cluster role binding", func() {
			rb := ResourceBuilder{
				Name:     "mesh-discovery-role-binding",
				Labels:   labels,
				Subjects: []rbacv1.Subject{GetServiceAccountSubject(namespace, name)},
				RoleRef:  GetClusterRoleRef(name),
			}
			testManifest.ExpectClusterRoleBinding(rb.GetClusterRoleBinding())
		})
	})
	Describe("custom resource", func() {

		var (
			gvk       = "MeshIngress"
			name      = "gloo"
			namespace = "supergloo-system"
		)

		It("has a crd with the given params", func() {
			testManifest.ExpectCustomResource(gvk, namespace, name)
		})

	})

	Describe("permissions", func() {
		It("has expected permissions associated with each deployment", func() {
			permissions := &ServiceAccountPermissions{}
			permissions.AddExpectedPermission(
				"sm-marketplace.mesh-discovery",
				"",
				[]string{""},
				[]string{"configmaps", "pods", "services", "secrets", "endpoints", "namespaces"},
				[]string{"get", "list", "watch"})
			permissions.AddExpectedPermission(
				"sm-marketplace.mesh-discovery",
				"",
				[]string{"authentication.istio.io"},
				[]string{"meshpolicies"},
				[]string{"get", "list", "watch"})
			permissions.AddExpectedPermission(
				"sm-marketplace.mesh-discovery",
				"",
				[]string{"apiextensions.k8s.io"},
				[]string{"customresourcedefinitions"},
				[]string{"get", "create"})
			// TODO this permissions stuff should get moved to the manifest tester
			testRbac := NewRbacTester(MustGetResources("example.yaml"))
			testRbac.ExpectServiceAccountPermissions(permissions)
		})
	})
})
