package testutils

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/solo-io/go-utils/testutils/kubectl"

	"github.com/onsi/ginkgo/v2"
	"github.com/pkg/errors"
)

// Deprecated: this function is incredibly slow, use CreateNamespacesInParallel instead
func SetupKubeForTest(namespace string) error {
	context := os.Getenv("KUBECTL_CONTEXT")
	if context == "" {
		current, err := KubectlOut("config", "current-context")
		if err != nil {
			return errors.Wrap(err, "getting currrent context")
		}
		context = strings.TrimSuffix(current, "\n")
	}
	// TODO(yuval-k): this changes the context for the user? can we do this less intrusive? maybe add it to
	// each kubectl command?
	if err := Kubectl("config", "set-context", context, "--namespace="+namespace); err != nil {
		return errors.Wrap(err, "setting context")
	}
	return Kubectl("create", "namespace", namespace)
}

// Deprecated: this function is incredibly slow, use DeleteNamespacesInParallelBlocking instead
func TeardownKube(namespace string) error {
	return Kubectl("delete", "namespace", namespace)
}

// Deprecated; use testutils/kubectl.DeleteCrd
func DeleteCrd(crd string) error {
	return kubectl.DeleteCrd(context.Background(), crd, kubectl.NewParams())
}

// Deprecated; use testutils/kubectl.Kubectl
func Kubectl(args ...string) error {
	p := kubectl.NewParams(args...)
	p.Stdout = ginkgo.GinkgoWriter
	p.Stderr = ginkgo.GinkgoWriter
	return kubectl.Kubectl(context.Background(), p)
}

// Deprecated; use testutils/kubectl.KubectlOut
func KubectlOut(args ...string) (string, error) {
	return kubectl.KubectlOut(context.Background(), kubectl.NewParams(args...))
}

// Deprecated; use testutils/kubectl.KubectlOutAsync
func KubectlOutAsync(args ...string) (io.Reader, chan struct{}, error) {
	ctx, cancel := context.WithCancel(context.Background())

	buf, err := kubectl.KubectlOutAsync(ctx, kubectl.NewParams(args...))

	done := make(chan struct{})
	go func() {
		select {
		case <-done:
			cancel()
		}
	}()

	return buf, done, err
}

// Deprecated; use testutils/kubectl.KubectlOutChan
func KubectlOutChan(r io.Reader, args ...string) (<-chan io.Reader, chan struct{}, error) {
	ctx, cancel := context.WithCancel(context.Background())

	p := kubectl.NewParams(args...)
	p.Stdin = r

	result, err := kubectl.KubectlOutChan(ctx, p)

	done := make(chan struct{})
	go func() {
		select {
		case <-done:
			cancel()
		}
	}()

	return result, done, err
}

// WaitPodsRunning waits for all pods to be running
// Deprecated; use testutils/kubectl.WaitPodsRunning
func WaitPodsRunning(ctx context.Context, interval time.Duration, namespace string, labels ...string) error {
	return kubectl.WaitPodsRunning(ctx, interval, namespace, kubectl.NewParams(), labels...)
}

// Deprecated; use testutils/kubectl.WaitPodStatus
func WaitPodStatus(ctx context.Context, interval time.Duration, namespace, label, status string, finished func(output string) bool) error {
	return kubectl.WaitPodStatus(ctx, interval, namespace, label, status, finished, kubectl.NewParams())
}

// Deprecated; use testutils/kubectl.KubeLogs
func KubeLogs(label string) string {
	return kubectl.KubeLogs(context.Background(), label, "", kubectl.NewParams())
}
