package testutils

import (
	"bytes"
	"context"
	"fmt"
	"github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"k8s.io/client-go/kubernetes"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/solo-io/go-utils/logger"

	"github.com/onsi/ginkgo"
	"github.com/pkg/errors"
)

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

func TeardownKube(namespace string) error {
	return Kubectl("delete", "namespace", namespace)
}

func DeleteCrd(crd string) error {
	return Kubectl("delete", "crd", crd)
}

func Kubectl(args ...string) error {
	cmd := exec.Command("kubectl", args...)
	cmd.Env = os.Environ()
	// disable DEBUG=1 from getting through to kube
	for i, pair := range cmd.Env {
		if strings.HasPrefix(pair, "DEBUG") {
			cmd.Env = append(cmd.Env[:i], cmd.Env[i+1:]...)
			break
		}
	}
	cmd.Stdout = ginkgo.GinkgoWriter
	cmd.Stderr = ginkgo.GinkgoWriter
	logger.Debugf("running: %s", strings.Join(cmd.Args, " "))
	return cmd.Run()
}

func KubectlOut(args ...string) (string, error) {
	cmd := exec.Command("kubectl", args...)
	cmd.Env = os.Environ()
	// disable DEBUG=1 from getting through to kube
	for i, pair := range cmd.Env {
		if strings.HasPrefix(pair, "DEBUG") {
			cmd.Env = append(cmd.Env[:i], cmd.Env[i+1:]...)
			break
		}
	}
	logger.Debugf("running: %s", strings.Join(cmd.Args, " "))
	out, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%s (%v)", out, err)
	}
	return string(out), err
}

func KubectlOutAsync(args ...string) (*bytes.Buffer, chan struct{}, error) {
	cmd := exec.Command("kubectl", args...)
	cmd.Env = os.Environ()
	// disable DEBUG=1 from getting through to kube
	for i, pair := range cmd.Env {
		if strings.HasPrefix(pair, "DEBUG") {
			cmd.Env = append(cmd.Env[:i], cmd.Env[i+1:]...)
			break
		}
	}
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	logger.Debugf("async running: %s", strings.Join(cmd.Args, " "))
	err := cmd.Start()
	if err != nil {
		err = fmt.Errorf("%s (%v)", buf.Bytes(), err)
	}
	done := make(chan struct{})
	go func() {
		select {
		case <-done:
			cmd.Process.Kill()
		}
	}()
	return buf, done, err
}

// WaitPodsRunning waits for all pods to be running
func WaitPodsRunning(ctx context.Context, interval time.Duration, namespace string, labels ...string) error {
	finished := func(output string) bool {
		return strings.Contains(output, "Running") || strings.Contains(output, "ContainerCreating")
	}
	for _, label := range labels {
		if err := WaitPodStatus(ctx, interval, namespace, label, "Running or ContainerCreating", finished); err != nil {
			return err
		}
	}
	finished = func(output string) bool {
		return strings.Contains(output, "Running")
	}
	for _, label := range labels {
		if err := WaitPodStatus(ctx, interval, namespace, label, "Running", finished); err != nil {
			return err
		}
	}
	return nil
}

func WaitPodStatus(ctx context.Context, interval time.Duration, namespace, label, status string, finished func(output string) bool) error {
	tick := time.Tick(interval)
	deadline, _ := ctx.Deadline()
	logger.Debugf("waiting till %v for pod %v to be %v...", deadline, label, status)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for %v to be %v", label, status)
		case <-tick:
			out, err := KubectlOut("get", "pod", "-l", label, "-n", namespace)
			if err != nil {
				return fmt.Errorf("failed getting pod: %v", err)
			}
			if strings.Contains(out, "CrashLoopBackOff") {
				out = KubeLogs(label)
				return errors.Errorf("%v in crash loop with logs %v", label, out)
			}
			if strings.Contains(out, "ErrImagePull") || strings.Contains(out, "ImagePullBackOff") {
				out, _ = KubectlOut("describe", "pod", "-l", label)
				return errors.Errorf("%v in ErrImagePull with description %v", label, out)
			}
			if finished(out) {
				return nil
			}
		}
	}
}

func KubeLogs(label string) string {
	out, err := KubectlOut("logs", "-l", label)
	if err != nil {
		out = err.Error()
	}
	return out
}

func MustKubeClient() kubernetes.Interface {
	restConfig, err := kubeutils.GetConfig("", "")
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	return kubeClient
}

