package helper

import (
	"bytes"

	"github.com/solo-io/go-utils/testutils"
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
	testContainer, err := newTestContainer(namespace, defaultTestRunnerImage, testrunnerName, TestRunnerPort)
	if err != nil {
		return nil, err
	}
	return &TestRunner{
		TestContainer: testContainer,
	}, nil
}

// This object represents a container that gets deployed to the cluster to support testing.
type TestRunner struct {
	*TestContainer
}

// TestContainer executes a command inside the TestContainer container
func (t *TestRunner) Exec(command ...string) (string, error) {
	args := append([]string{"exec", "-i", t.echoName, "-n", t.namespace, "--"}, command...)
	return testutils.KubectlOut(args...)
}

// TestRunnerAsync executes a command inside the TestContainer container
// returning a buffer that can be read from as it executes
func (t *TestRunner) TestRunnerAsync(args ...string) (*bytes.Buffer, chan struct{}, error) {
	args = append([]string{"exec", "-i", t.echoName, "-n", t.namespace, "--"}, args...)
	return testutils.KubectlOutAsync(args...)
}
