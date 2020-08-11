package kube

import (
	"context"
	"log"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func WaitForNamespaceTeardown(ctx context.Context, ns string) {
	WaitForNamespaceTeardownWithClient(ctx, ns, MustKubeClient())
}

func WaitForNamespaceTeardownWithClient(ctx context.Context, ns string, client kubernetes.Interface) {
	EventuallyWithOffset(1, func() (bool, error) {
		namespaces, err := client.CoreV1().Namespaces().List(ctx, v1.ListOptions{})
		if err != nil {
			// namespace is gone
			return false, err
		}
		for _, n := range namespaces.Items {
			if n.Name == ns {
				return false, nil
			}
		}
		return true, nil
	}, time.Second*180).Should(BeTrue())
}

func WaitUntilPodsRunning(ctx context.Context, timeout time.Duration, namespace string, podPrefixes ...string) error {
	pods := MustKubeClient().CoreV1().Pods(namespace)
	podsWithPrefixReady := func(prefix string) (bool, error) {
		list, err := pods.List(ctx, metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		var podsWithPrefix []corev1.Pod
		for _, pod := range list.Items {
			if strings.HasPrefix(pod.Name, prefix) {
				podsWithPrefix = append(podsWithPrefix, pod)
			}
		}
		if len(podsWithPrefix) == 0 {
			return false, eris.Errorf("no pods found with prefix %v", prefix)
		}
		for _, pod := range podsWithPrefix {
			var podReady bool
			for _, cond := range pod.Status.Conditions {
				if cond.Type == corev1.ContainersReady && cond.Status == corev1.ConditionTrue {
					podReady = true
					break
				}
			}
			if !podReady {
				return false, nil
			}
		}
		return true, nil
	}
	failed := time.After(timeout)
	notYetRunning := make(map[string]struct{})
	for {
		select {
		case <-failed:
			return eris.Errorf("timed out waiting for pods to come online: %v", notYetRunning)
		case <-time.After(time.Second / 2):
			notYetRunning = make(map[string]struct{})
			for _, prefix := range podPrefixes {
				ready, err := podsWithPrefixReady(prefix)
				if err != nil {
					log.Printf("failed to get pod status: %v", err)
					notYetRunning[prefix] = struct{}{}
				}
				if !ready {
					notYetRunning[prefix] = struct{}{}
				}
			}
			if len(notYetRunning) == 0 {
				return nil
			}
		}

	}
}
