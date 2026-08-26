package testutils

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/go-utils/threadsafe"

	"github.com/onsi/ginkgo/v2"
	"github.com/pkg/errors"
)

// Helper vars that allow us to mock the kubectl call with static echo text
var (
	kubectlExecutable          = "kubectl"
	kubectlArgs       []string = nil
)

// Deprecated: this function is incredibly slow, use CreateNamespacesInParallel instead
func SetupKubeForTest(namespace string) error {
	context := os.Getenv("KUBECTL_CONTEXT")
	if context == "" {
		current, err := KubectlOut("config", "current-context")
		if err != nil {
			return errors.Wrap(err, "getting current context")
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

func DeleteCrd(crd string) error {
	return Kubectl("delete", "crd", crd)
}

func kubectl(args ...string) *exec.Cmd {
	if kubectlArgs != nil {
		args = kubectlArgs
	}
	cmd := exec.Command(kubectlExecutable, args...)
	cmd.Env = os.Environ()
	// disable DEBUG=1 from getting through to kube
	for i, pair := range cmd.Env {
		if strings.HasPrefix(pair, "DEBUG") {
			cmd.Env = append(cmd.Env[:i], cmd.Env[i+1:]...)
			break
		}
	}
	return cmd
}

func Kubectl(args ...string) error {
	cmd := kubectl(args...)
	cmd.Stdout = ginkgo.GinkgoWriter
	cmd.Stderr = ginkgo.GinkgoWriter
	log.Debugf("running: %s", strings.Join(cmd.Args, " "))
	return cmd.Run()
}

func KubectlOut(args ...string) (string, error) {
	cmd := kubectl(args...)
	log.Debugf("running: %s", strings.Join(cmd.Args, " "))
	out, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%s (%v)", out, err)
	}
	return string(out), err
}

func KubectlOutAsync(args ...string) (io.Reader, chan struct{}, error) {
	cmd := kubectl(args...)
	buf := &threadsafe.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	log.Debugf("async running: %s", strings.Join(cmd.Args, " "))
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

func KubectlOutChan(r io.Reader, args ...string) (<-chan io.Reader, chan struct{}, error) {
	cmd := kubectl(args...)
	buf := &threadsafe.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	cmd.Stdin = r
	log.Debugf("async running: %s", strings.Join(cmd.Args, " "))
	err := cmd.Start()
	if err != nil {
		return nil, nil, err
	}
	done := make(chan struct{})
	go func() {
		select {
		case <-done:
			cmd.Process.Kill()
		}
	}()

	result := make(chan io.Reader)
	go func() {
		for {
			select {
			case <-time.After(time.Second):
				select {
				case result <- buf:
					continue
				case <-done:
					return
				default:
					continue
				}
			}
		}
	}()

	return result, done, err
}

// WaitPodsRunning waits for all pods to be running
func WaitPodsRunning(ctx context.Context, interval time.Duration, namespace string, labels ...string) error {
	finished := func(output string) bool {
		return strings.Contains(output, "Running") && !strings.Contains(output, "Terminating")
	}
	for _, label := range labels {
		if err := WaitPodStatus(ctx, interval, namespace, label, "Running, not Terminating", finished); err != nil {
			return err
		}
	}
	return nil
}

func WaitPodStatus(ctx context.Context, interval time.Duration, namespace, label, status string, finished func(output string) bool) error {
	tick := time.Tick(interval)
	deadline, _ := ctx.Deadline()
	log.Debugf("waiting till %v for pod %v to be %v...", deadline, label, status)
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
				return eris.Errorf("%v in crash loop with logs %v", label, out)
			}
			if strings.Contains(out, "ErrImagePull") || strings.Contains(out, "ImagePullBackOff") {
				out, _ = KubectlOut("describe", "pod", "-l", label)
				return eris.Errorf("%v in ErrImagePull with description %v", label, out)
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
