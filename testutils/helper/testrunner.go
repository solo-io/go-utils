package helper

import (
	"fmt"
	"time"

	"github.com/solo-io/go-utils/log"
)

const (
	defaultTestRunnerImage = "soloio/testrunner:latest"
	TestrunnerName         = "testrunner"
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

func NewTestRunner(namespace string) (*testRunner, error) {
	testContainer, err := newTestContainer(namespace, defaultTestRunnerImage, TestrunnerName, TestRunnerPort)
	if err != nil {
		return nil, err
	}

	return &testRunner{
		testContainer: testContainer,
	}, nil
}

// This object represents a container that gets deployed to the cluster to support testing.
type testRunner struct {
	*testContainer
}

func (t *testRunner) Deploy(timeout time.Duration) error {
	err := t.deploy(timeout)
	if err != nil {
		return err
	}
	go func() {
		start := time.Now()
		log.Debugf("starting http server listening on port %v", TestRunnerPort)
		// This command start an http SimpleHttpServer and blocks until the server terminates
		if _, err := t.Exec("python", "-m", "SimpleHTTPServer", fmt.Sprintf("%v", TestRunnerPort)); err != nil {
			// if an error happened after 5 seconds, it's probably not an error.. just the pod terminating.
			if time.Now().Sub(start).Seconds() < 5.0 {
				log.Warnf("failed to start HTTP Server in Test Runner: %v", err)
			}
		}
	}()
	return nil
}
