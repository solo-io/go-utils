package kube

import (
	"context"
	"os"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/kubeutils"
	"go.uber.org/zap"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateNs(ns string) error {
	kube := MustKubeClient()
	_, err := kube.CoreV1().Namespaces().Create(context.Background(), &kubev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
		},
	}, metav1.CreateOptions{})

	return err
}

func MustCreateNs(ns string) {
	ExpectWithOffset(1, CreateNs(ns)).NotTo(HaveOccurred())
}

func DeleteNs(ns string) error {
	kube := MustKubeClient()
	err := kube.CoreV1().Namespaces().Delete(context.Background(), ns, metav1.DeleteOptions{})

	return err
}

func MustDeleteNs(ns string) {
	ExpectWithOffset(1, DeleteNs(ns)).NotTo(HaveOccurred())
}

func ConfigMap(ns, name, data string, labels map[string]string) kubev1.ConfigMap {
	return kubev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    labels,
		},
		Data: map[string]string{"data": data},
	}
}

func CreateConfigMap(cm kubev1.ConfigMap) error {
	kube := MustKubeClient()
	_, err := kube.CoreV1().ConfigMaps(cm.Namespace).Create(context.Background(), &cm, metav1.CreateOptions{})

	return err
}

func MustCreateConfigMap(cm kubev1.ConfigMap) {
	ExpectWithOffset(1, CreateConfigMap(cm)).NotTo(HaveOccurred())
}

func MustKubeClient() kubernetes.Interface {
	client, err := KubeClient()
	if err != nil {
		contextutils.LoggerFrom(context.TODO()).Fatalw("failed to create kube client", zap.Error(err))
	}
	return client
}

func KubeClient() (kubernetes.Interface, error) {
	cfg, err := kubeutils.GetConfig("", os.Getenv("KUBECONFIG"))
	if err != nil {
		return nil, errors.Wrapf(err, "getting kube config")
	}
	return kubernetes.NewForConfig(cfg)
}
