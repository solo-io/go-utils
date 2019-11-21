package kuberesource

import (
	"sort"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// adapted from Helm 2
// https://github.com/helm/helm/blob/release-2.16/pkg/tiller/kind_sorter.go#L113

// InstallOrder is the order in which manifests should be installed (by Kind).
//
// Those occurring earlier in the list get installed before those occurring later in the list.
var InstallOrder = []string{
	"Namespace",
	"NetworkPolicy",
	"ResourceQuota",
	"LimitRange",
	"PodSecurityPolicy",
	"PodDisruptionBudget",
	"Secret",
	"ConfigMap",
	"StorageClass",
	"PersistentVolume",
	"PersistentVolumeClaim",
	"ServiceAccount",
	"CustomResourceDefinition",
	"ClusterRole",
	"ClusterRoleList",
	"ClusterRoleBinding",
	"ClusterRoleBindingList",
	"Role",
	"RoleList",
	"RoleBinding",
	"RoleBindingList",
	"Service",
	"DaemonSet",
	"Pod",
	"ReplicationController",
	"ReplicaSet",
	"Deployment",
	"HorizontalPodAutoscaler",
	"StatefulSet",
	"Job",
	"CronJob",
	"Ingress",
	"APIService",
}

var customInstallOrder = func() []string {
	// insert MutatingWebhookConfiguration after Namespace
	return append(InstallOrder[:1], append([]string{"MutatingWebhookConfiguration"}, InstallOrder[1:]...)...)
}()

// set the install order based on
// Helm's list
var installOrder = func() map[string]int {
	installOrderMapping := make(map[string]int)
	for installValue, kind := range customInstallOrder {
		installOrderMapping[kind] = installValue + 1 // add 1, the 0 value is unknown type
	}
	return installOrderMapping
}()

func getInstallOrder(res *unstructured.Unstructured) (int, bool) {
	kind := res.GroupVersionKind().Kind

	order, listed := installOrder[kind]
	return order, listed
}

func sortUnstructured(resources UnstructuredResources) {
	sort.SliceStable(resources, func(i, j int) bool {
		res1 := resources[i]
		res2 := resources[j]

		installOrder1, listed1 := getInstallOrder(res1)
		installOrder2, listed2 := getInstallOrder(res2)
		kind1, kind2 := res1.GroupVersionKind().Kind, res2.GroupVersionKind().Kind

		// unlisted objects come last
		switch {
		case !listed1 && !listed2:
			if kind1 != kind2 {
				return kind1 < kind2
			}
			return nameLess(res1, res2)
		case !listed1:
			return false
		case !listed2:
			return true
		}

		if installOrder1 != installOrder2 {
			return installOrder1 < installOrder2
		}

		if kind1 != kind2 {
			return kind1 < kind2
		}

		// sort by namespace.name
		val := nameLess(res1, res2)
		return val
	})
}

func nameLess(res1, res2 *unstructured.Unstructured) bool {
	return res1.GetNamespace()+res1.GetName() < res2.GetNamespace()+res2.GetName()
}

func installOrderLess(kind1, kind2 string) bool {
	order1, listed1 := installOrder[kind1]
	order2, listed2 := installOrder[kind2]
	switch {
	case !listed1 && !listed2:
		return true
	case !listed1:
		return false
	case !listed2:
		return true
	}

	return order1 < order2
}
