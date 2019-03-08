package install

import (
	"fmt"
	"strings"
	"time"

	"github.com/onsi/gomega"
	"github.com/solo-io/go-utils/logger"
)

type CurlOpts struct {
	Protocol          string
	Path              string
	Method            string
	Host              string
	Service           string
	CaFile            string
	Body              string
	Headers           map[string]string
	Port              int
	ReturnHeaders     bool
	ConnectionTimeout int
}

func (t *TestRunner) CurlEventuallyShouldRespond(opts CurlOpts, substr string, ginkgoOffset int, timeout time.Duration) {
	defaultTimeout := time.Second * 20
	if timeout == 0 {
		timeout = defaultTimeout
	}
	// for some useful-ish output
	tick := time.Tick(timeout / 8)

	gomega.EventuallyWithOffset(ginkgoOffset, func() string {
		res, err := t.Curl(opts)
		if err != nil {
			res = err.Error()
			// trigger an early exit if the pod has been deleted
			gomega.Expect(res).NotTo(gomega.ContainSubstring(`pods "testrunner" not found`))
		}
		select {
		default:
			break
		case <-tick:
			logger.GreyPrintf("running: %v\nwant %v\nhave: %s", opts, substr, res)
		}
		if strings.Contains(res, substr) {
			logger.GreyPrintf("success: %v", res)
		}
		return res
	}, t, "5s").Should(gomega.ContainSubstring(substr))
}

func (t *TestRunner) Curl(opts CurlOpts) (string, error) {
	args := []string{"curl", "-v"}
	if opts.ConnectionTimeout > 0 {
		seconds := fmt.Sprintf("%v", opts.ConnectionTimeout)
		args = append(args, "--connect-timeout", seconds, "--max-time", seconds)
	}

	if opts.ReturnHeaders {
		args = append(args, "-I")
	}

	if opts.Method != "GET" && opts.Method != "" {
		args = append(args, "-X"+opts.Method)
	}
	if opts.Host != "" {
		args = append(args, "-H", "Host: "+opts.Host)
	}
	if opts.CaFile != "" {
		args = append(args, "--cacert", opts.CaFile)
	}
	if opts.Body != "" {
		args = append(args, "-H", "Content-Type: application/json")
		args = append(args, "-d", opts.Body)
	}
	for h, v := range opts.Headers {
		args = append(args, "-H", fmt.Sprintf("%v: %v", h, v))
	}
	port := opts.Port
	if port == 0 {
		port = 8080
	}
	protocol := opts.Protocol
	if protocol == "" {
		protocol = "http"
	}
	service := opts.Service
	if service == "" {
		service = "test-ingress"
	}
	args = append(args, fmt.Sprintf("%v://%s:%v%s", protocol, service, port, opts.Path))
	logger.Debugf("running: curl %v", strings.Join(args, " "))
	return t.Exec(args...)
}
