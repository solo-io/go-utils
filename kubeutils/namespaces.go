package kubeutils

import (
	"context"

	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateNamespacesInParallel(ctx context.Context, kube kubernetes.Interface, namespaces ...string) error {
	eg := errgroup.Group{}
	for _, namespace := range namespaces {
		namespace := namespace
		eg.Go(func() error {
			_, err := kube.CoreV1().Namespaces().Create(ctx, &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}, metav1.CreateOptions{})
			return err
		})
	}
	return eg.Wait()
}

func DeleteNamespacesInParallelBlocking(ctx context.Context, kube kubernetes.Interface, namespaces ...string) error {
	eg := errgroup.Group{}
	for _, namespace := range namespaces {
		namespace := namespace
		eg.Go(func() error {
			return kube.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
		})
	}
	return eg.Wait()
}

func DeleteNamespacesInParallel(ctx context.Context, kube kubernetes.Interface, namespaces ...string) {
	for _, namespace := range namespaces {
		namespace := namespace
		go func() {
			kube.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
		}()
	}
}
