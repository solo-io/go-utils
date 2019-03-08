package install

import (
	"context"
	"fmt"
	"github.com/prometheus/common/log"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/testutils"
	"strings"
	"time"

	"github.com/solo-io/go-utils/errors"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	defaultTestRunnerImage = "soloio/testrunner:latest"
	TestRunnerPort         = 1234
)

func DeployTestRunner(namespace string) error {
	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return err
	}
	kube, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}
	labels := map[string]string{"gloo": "testrunner"}
	zero := int64(0)
	if _, err := kube.CoreV1().Pods(namespace).Create(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testrunner",
			Namespace: namespace,
			// needed for waitPodsRunning
			Labels: labels,
		},
		Spec: v1.PodSpec{
			TerminationGracePeriodSeconds: &zero,
			Containers: []v1.Container{
				{
					Image:           defaultTestRunnerImage,
					ImagePullPolicy: v1.PullIfNotPresent,
					Name:            "testrunner",
				},
			},
		},
	}); err != nil {
		return err
	}
	if _, err := kube.CoreV1().Services(namespace).Create(&v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testrunner",
			Namespace: namespace,
			// needed for waitPodsRunning
			Labels: labels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:     "http",
					Protocol: v1.ProtocolTCP,
					Port:     TestRunnerPort,
				},
			},
			Selector: labels,
		},
	}); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := waitPodsRunning(ctx, time.Second, namespace, "gloo=testrunner"); err != nil {
		return err
	}
	go func() {
		start := time.Now()
		if err := startSimpleHttpServer(namespace, TestRunnerPort); err != nil {
			// if an error happened after 5 seconds, it's probably not an error.. just the pod terminating.
			if time.Now().Sub(start).Seconds() < 5.0 {
				log.Warnf("failed to start HTTP Server in Test Runner: %v", err)
			}
		}
	}()
	return nil
}

func waitPodsRunning(ctx context.Context, interval time.Duration, namespace string, labels ...string) error {
	finished := func(output string) bool {
		return strings.Contains(output, "Running") || strings.Contains(output, "ContainerCreating")
	}
	for _, label := range labels {
		if err := waitPodStatus(ctx, interval, namespace, label, "Running or ContainerCreating", finished); err != nil {
			return err
		}
	}
	finished = func(output string) bool {
		return strings.Contains(output, "Running")
	}
	for _, label := range labels {
		if err := waitPodStatus(ctx, interval, namespace, label, "Running", finished); err != nil {
			return err
		}
	}
	return nil
}

func waitPodStatus(ctx context.Context, interval time.Duration, namespace, label, status string, finished func(output string) bool) error {
	tick := time.Tick(interval)
	d, _ := ctx.Deadline()
	log.Debugf("waiting till %v for pod %v to be %v...", d, label, status)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for %v to be %v", label, status)
		case <-tick:
			out, err := testutils.KubectlOut("get", "pod", "-l", label, "-n", namespace)
			if err != nil {
				return fmt.Errorf("failed getting pod: %v", err)
			}
			if strings.Contains(out, "CrashLoopBackOff") {
				out = KubeLogs(label)
				return errors.Errorf("%v in crash loop with logs %v", label, out)
			}
			if strings.Contains(out, "ErrImagePull") || strings.Contains(out, "ImagePullBackOff") {
				out, _ = testutils.KubectlOut("describe", "pod", "-l", label)
				return errors.Errorf("%v in ErrImagePull with description %v", label, out)
			}
			if finished(out) {
				return nil
			}
		}
	}
}

func KubeLogs(label string) string {
	out, err := testutils.KubectlOut("logs", "-l", label)
	if err != nil {
		out = err.Error()
	}
	return out
}

// this response is given by the testrunner when the SimpleServer is started
const SimpleHttpResponse = `<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 3.2 Final//EN"><html>
<title>Directory listing for /</title>
<body>
<h2>Directory listing for /</h2>
<hr>
<ul>
<li><a href="bin/">bin/</a>
<li><a href="pkg/">pkg/</a>
<li><a href="protoc-3.3.0-linux-x86_64.zip">protoc-3.3.0-linux-x86_64.zip</a>
<li><a href="protoc3/">protoc3/</a>
<li><a href="src/">src/</a>
</ul>
<hr>
</body>
</html>`

func startSimpleHttpServer(namespace string, port int32) error {
	_, err := testutils.TestRunner(namespace, "python", "-m", "SimpleHTTPServer", fmt.Sprintf("%v", port))
	return err
}
