package kube

import (
	"context"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	kubev1 "k8s.io/api/core/v1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func WaitForServicesInNamespaceTeardown(ns string) {
	EventuallyWithOffset(1, func() []kubev1.Service {
		svcs, err := MustKubeClient().CoreV1().Services(ns).List(context.Background(), v1.ListOptions{})
		if err != nil {
			// namespace is gone
			return []kubev1.Service{}
		}
		return svcs.Items
	}, time.Second*30).Should(BeEmpty())
}

func TeardownClusterResourcesWithPrefix(kube kubernetes.Interface, prefix string) {
	clusterroles, err := kube.RbacV1beta1().ClusterRoles().List(context.Background(), metav1.ListOptions{})
	if err == nil {
		for _, cr := range clusterroles.Items {
			if strings.Contains(cr.Name, prefix) {
				kube.RbacV1beta1().ClusterRoles().Delete(context.Background(), cr.Name, v1.DeleteOptions{})
			}
		}
	}
	clusterrolebindings, err := kube.RbacV1beta1().ClusterRoleBindings().List(context.Background(), metav1.ListOptions{})
	if err == nil {
		for _, cr := range clusterrolebindings.Items {
			if strings.Contains(cr.Name, prefix) {
				kube.RbacV1beta1().ClusterRoleBindings().Delete(context.Background(), cr.Name, v1.DeleteOptions{})
			}
		}
	}
	webhooks, err := kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().List(context.Background(), metav1.ListOptions{})
	if err == nil {
		for _, wh := range webhooks.Items {
			if strings.Contains(wh.Name, prefix) {
				kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Delete(context.Background(), wh.Name, v1.DeleteOptions{})
			}
		}
	}

	cfg, err := kubeutils.GetConfig("", "")
	Expect(err).NotTo(HaveOccurred())

	exts, err := apiexts.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())

	crds, err := exts.ApiextensionsV1beta1().CustomResourceDefinitions().List(context.Background(), metav1.ListOptions{})
	if err == nil {
		for _, cr := range crds.Items {
			if strings.Contains(cr.Name, prefix) {
				exts.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(context.Background(), cr.Name, v1.DeleteOptions{})
			}
		}
	}
}
