package install

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/solo-io/go-utils/logger"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/testutils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	defaultTestRunnerImage = "soloio/testrunner:latest"
	testrunnerName         = "testrunner"
	TestRunnerPort         = 1234

	// This response is given by the testrunner when the SimpleServer is started
	SimpleHttpResponse = `<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 3.2 Final//EN"><html>
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
)

func NewTestRunner(namespace string) (*TestRunner, error) {
	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, err
	}
	kube, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &TestRunner{
		namespace: namespace,
		kube:      kube,
	}, nil
}

// This object represents a container that gets deployed to the cluster to support testing.
type TestRunner struct {
	containerImageName string
	containerPort      uint
	namespace          string
	kube               *kubernetes.Clientset
}

// Deploys the test runner to the kubernetes cluster the kubeconfig is pointing to and waits for the given time for the
// testrunner pod to be running.
func (t *TestRunner) Deploy(timeout time.Duration) error {
	zero := int64(0)
	labels := map[string]string{"gloo": testrunnerName}
	metadata := metav1.ObjectMeta{
		Name:      testrunnerName,
		Namespace: t.namespace,
		Labels:    labels,
	}

	// Create test runner pod
	if _, err := t.kube.CoreV1().Pods(t.namespace).Create(&corev1.Pod{
		ObjectMeta: metadata,
		Spec: corev1.PodSpec{
			TerminationGracePeriodSeconds: &zero,
			Containers: []corev1.Container{
				{
					Image:           defaultTestRunnerImage,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Name:            testrunnerName,
				},
			},
		},
	}); err != nil {
		return err
	}

	// Create test runner service
	if _, err := t.kube.CoreV1().Services(t.namespace).Create(&corev1.Service{
		ObjectMeta: metadata,
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Protocol: corev1.ProtocolTCP,
					Port:     TestRunnerPort,
				},
			},
			Selector: labels,
		},
	}); err != nil {
		return err
	}

	// Wait until the test runner pod is running
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := testutils.WaitPodsRunning(ctx, time.Second, t.namespace, "gloo="+testrunnerName); err != nil {
		return err
	}

	go func() {
		start := time.Now()
		logger.Debugf("starting http server listening on port %v", TestRunnerPort)
		// This command start an http SimpleHttpServer and blocks until the server terminates
		if _, err := t.Exec("python", "-m", "SimpleHTTPServer", fmt.Sprintf("%v", TestRunnerPort)); err != nil {
			// if an error happened after 5 seconds, it's probably not an error.. just the pod terminating.
			if time.Now().Sub(start).Seconds() < 5.0 {
				logger.Warnf("failed to start HTTP Server in Test Runner: %v", err)
			}
		}
	}()
	return nil
}

func (t *TestRunner) Terminate() error {
	if err := testutils.Kubectl("delete", "pod", "-n", t.namespace, testrunnerName, "--grace-period=0"); err != nil {
		return errors.Wrapf(err, "deleting %s pod", testrunnerName)
	}
	return nil

}

// TestRunner executes a command inside the TestRunner container
func (t *TestRunner) Exec(command ...string) (string, error) {
	args := append([]string{"exec", "-i", testrunnerName, "-n", t.namespace, "--"}, command...)
	return testutils.KubectlOut(args...)
}

// TestRunnerAsync executes a command inside the TestRunner container
// returning a buffer that can be read from as it executes
func (t *TestRunner) TestRunnerAsync(args ...string) (*bytes.Buffer, chan struct{}, error) {
	args = append([]string{"exec", "-i", testrunnerName, "-n", t.namespace, "--"}, args...)
	return testutils.KubectlOutAsync(args...)
}
