package kubeinstallutils

import (
	"context"

	"github.com/rotisserie/eris"
	"k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	autoscaling "k8s.io/api/autoscaling/v1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	rbac "k8s.io/api/rbac/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

type KubeInstaller interface {
	Create(ctx context.Context, obj runtime.Object) error
	Update(ctx context.Context, obj runtime.Object) error
	Delete(ctx context.Context, obj runtime.Object) error
}

// generic kube installer, CUD arbitrary kube objects
func NewKubeInstaller(kube kubernetes.Interface, exts apiexts.Interface, namespace string) KubeInstaller {
	return &kubeInstaller{
		kube:      kube,
		exts:      exts,
		namespace: namespace,
	}
}

type kubeInstaller struct {
	kube      kubernetes.Interface
	exts      apiexts.Interface
	namespace string
}

func (k *kubeInstaller) Create(ctx context.Context, obj runtime.Object) error {
	kube := k.kube
	exts := k.exts
	namespace := k.namespace
	type namespaceable interface {
		GetNamespace() string
	}
	if namespacedObj, ok := obj.(namespaceable); ok && namespace == "" {
		namespace = namespacedObj.GetNamespace()
	}
	switch obj := obj.(type) {
	case *core.Namespace:
		_, err := kube.CoreV1().Namespaces().Create(ctx, obj, metav1.CreateOptions{})
		return err
	case *core.ConfigMap:
		_, err := kube.CoreV1().ConfigMaps(namespace).Create(ctx, obj, metav1.CreateOptions{})
		return err
	case *core.ServiceAccount:
		_, err := kube.CoreV1().ServiceAccounts(namespace).Create(ctx, obj, metav1.CreateOptions{})
		return err
	case *core.Service:
		_, err := kube.CoreV1().Services(namespace).Create(ctx, obj, metav1.CreateOptions{})
		return err
	case *core.Pod:
		_, err := kube.CoreV1().Pods(namespace).Create(ctx, obj, metav1.CreateOptions{})
		return err
	case *rbac.ClusterRole:
		_, err := kube.RbacV1().ClusterRoles().Create(ctx, obj, metav1.CreateOptions{})
		return err
	case *rbac.ClusterRoleBinding:
		_, err := kube.RbacV1().ClusterRoleBindings().Create(ctx, obj, metav1.CreateOptions{})
		return err
	case *batch.Job:
		_, err := kube.BatchV1().Jobs(namespace).Create(ctx, obj, metav1.CreateOptions{})
		return err
	case *appsv1beta2.Deployment:
		_, err := kube.AppsV1beta2().Deployments(namespace).Create(ctx, obj, metav1.CreateOptions{})
		return err
	case *appsv1.Deployment:
		_, err := kube.AppsV1().Deployments(namespace).Create(ctx, obj, metav1.CreateOptions{})
		return err
	case *appsv1beta2.DaemonSet:
		_, err := kube.AppsV1beta2().DaemonSets(namespace).Create(ctx, obj, metav1.CreateOptions{})
		return err
	case *appsv1.DaemonSet:
		_, err := kube.AppsV1().DaemonSets(namespace).Create(ctx, obj, metav1.CreateOptions{})
		return err
	case *extensionsv1beta1.Deployment:
		_, err := kube.ExtensionsV1beta1().Deployments(namespace).Create(ctx, obj, metav1.CreateOptions{})
		return err
	case *extensionsv1beta1.DaemonSet:
		_, err := kube.ExtensionsV1beta1().DaemonSets(namespace).Create(ctx, obj, metav1.CreateOptions{})
		return err
	case *apiextensions.CustomResourceDefinition:
		_, err := exts.ApiextensionsV1beta1().CustomResourceDefinitions().Create(ctx, obj, metav1.CreateOptions{})
		return err
	case *v1beta1.MutatingWebhookConfiguration:
		_, err := kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Create(ctx, obj, metav1.CreateOptions{})
		return err
	case *autoscaling.HorizontalPodAutoscaler:
		_, err := kube.AutoscalingV1().HorizontalPodAutoscalers(namespace).Create(ctx, obj, metav1.CreateOptions{})
		return err
	}
	return eris.Errorf("no implementation for type %v", obj)
}

// resource version should be ignored / not matter
func (k *kubeInstaller) Update(ctx context.Context, obj runtime.Object) error {
	kube := k.kube
	exts := k.exts
	namespace := k.namespace
	type namespaceable interface {
		GetNamespace() string
	}
	if namespacedObj, ok := obj.(namespaceable); ok && namespace == "" {
		namespace = namespacedObj.GetNamespace()
	}
	switch obj := obj.(type) {
	case *core.Namespace:
		client := kube.CoreV1().Namespaces()
		obj2, err := client.Get(ctx, obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(ctx, obj, metav1.UpdateOptions{})
		return err
	case *core.ConfigMap:
		client := kube.CoreV1().ConfigMaps(namespace)
		obj2, err := client.Get(ctx, obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(ctx, obj, metav1.UpdateOptions{})
		return err
	case *core.ServiceAccount:
		client := kube.CoreV1().ServiceAccounts(namespace)
		obj2, err := client.Get(ctx, obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(ctx, obj, metav1.UpdateOptions{})
		return err
	case *core.Service:
		client := kube.CoreV1().Services(namespace)
		obj2, err := client.Get(ctx, obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(ctx, obj, metav1.UpdateOptions{})
		return err
	case *core.Pod:
		client := kube.CoreV1().Pods(namespace)
		obj2, err := client.Get(ctx, obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(ctx, obj, metav1.UpdateOptions{})
		return err
	case *rbac.ClusterRole:
		client := kube.RbacV1().ClusterRoles()
		obj2, err := client.Get(ctx, obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(ctx, obj, metav1.UpdateOptions{})
		return err
	case *rbac.ClusterRoleBinding:
		client := kube.RbacV1().ClusterRoleBindings()
		obj2, err := client.Get(ctx, obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(ctx, obj, metav1.UpdateOptions{})
		return err
	case *batch.Job:
		client := kube.BatchV1().Jobs(namespace)
		obj2, err := client.Get(ctx, obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(ctx, obj, metav1.UpdateOptions{})
		return err
	case *appsv1beta2.Deployment:
		client := kube.AppsV1beta2().Deployments(namespace)
		obj2, err := client.Get(ctx, obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(ctx, obj, metav1.UpdateOptions{})
		return err
	case *appsv1beta2.DaemonSet:
		client := kube.AppsV1beta2().DaemonSets(namespace)
		obj2, err := client.Get(ctx, obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(ctx, obj, metav1.UpdateOptions{})
		return err
	case *appsv1.Deployment:
		client := kube.AppsV1().Deployments(namespace)
		obj2, err := client.Get(ctx, obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(ctx, obj, metav1.UpdateOptions{})
		return err
	case *appsv1.DaemonSet:
		client := kube.AppsV1().DaemonSets(namespace)
		obj2, err := client.Get(ctx, obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(ctx, obj, metav1.UpdateOptions{})
		return err
	case *apiextensions.CustomResourceDefinition:
		client := exts.ApiextensionsV1beta1().CustomResourceDefinitions()
		obj2, err := client.Get(ctx, obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(ctx, obj, metav1.UpdateOptions{})
		return err
	case *v1beta1.MutatingWebhookConfiguration:
		client := kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations()
		obj2, err := client.Get(ctx, obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(ctx, obj, metav1.UpdateOptions{})
		return err
	case *autoscaling.HorizontalPodAutoscaler:
		client := kube.AutoscalingV1().HorizontalPodAutoscalers(namespace)
		obj2, err := client.Get(ctx, obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(ctx, obj, metav1.UpdateOptions{})
		return err
	}
	return eris.Errorf("no implementation for type %v", obj)
}

// this can be just an empty object of the correct type w/ the name and namespace (if applicable) set
func (k *kubeInstaller) Delete(ctx context.Context, obj runtime.Object) error {
	kube := k.kube
	exts := k.exts
	namespace := k.namespace
	type namespaceable interface {
		GetNamespace() string
	}
	if namespacedObj, ok := obj.(namespaceable); ok && namespace == "" {
		namespace = namespacedObj.GetNamespace()
	}
	switch obj := obj.(type) {
	case *core.Namespace:
		return kube.CoreV1().Namespaces().Delete(ctx, obj.Name, metav1.DeleteOptions{})
	case *core.ConfigMap:
		return kube.CoreV1().ConfigMaps(namespace).Delete(ctx, obj.Name, metav1.DeleteOptions{})
	case *core.ServiceAccount:
		return kube.CoreV1().ServiceAccounts(namespace).Delete(ctx, obj.Name, metav1.DeleteOptions{})
	case *core.Service:
		return kube.CoreV1().Services(namespace).Delete(ctx, obj.Name, metav1.DeleteOptions{})
	case *core.Pod:
		return kube.CoreV1().Pods(namespace).Delete(ctx, obj.Name, metav1.DeleteOptions{})
	case *rbac.ClusterRole:
		return kube.RbacV1().ClusterRoles().Delete(ctx, obj.Name, metav1.DeleteOptions{})
	case *rbac.ClusterRoleBinding:
		return kube.RbacV1().ClusterRoleBindings().Delete(ctx, obj.Name, metav1.DeleteOptions{})
	case *batch.Job:
		return kube.BatchV1().Jobs(namespace).Delete(ctx, obj.Name, metav1.DeleteOptions{})
	case *appsv1.Deployment:
		return kube.AppsV1().Deployments(namespace).Delete(ctx, obj.Name, metav1.DeleteOptions{})
	case *appsv1.DaemonSet:
		return kube.AppsV1().DaemonSets(namespace).Delete(ctx, obj.Name, metav1.DeleteOptions{})
	case *appsv1beta2.Deployment:
		return kube.AppsV1beta2().Deployments(namespace).Delete(ctx, obj.Name, metav1.DeleteOptions{})
	case *appsv1beta2.DaemonSet:
		return kube.AppsV1beta2().DaemonSets(namespace).Delete(ctx, obj.Name, metav1.DeleteOptions{})
	case *apiextensions.CustomResourceDefinition:
		return exts.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(ctx, obj.Name, metav1.DeleteOptions{})
	case *v1beta1.MutatingWebhookConfiguration:
		return kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Delete(ctx, obj.Name, metav1.DeleteOptions{})
	case *autoscaling.HorizontalPodAutoscaler:
		return kube.AutoscalingV1().HorizontalPodAutoscalers(namespace).Delete(ctx, obj.Name, metav1.DeleteOptions{})
	}
	return eris.Errorf("no implementation for type %v", obj)
}
