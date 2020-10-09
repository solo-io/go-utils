package cliutils

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/hashicorp/go-multierror"
	errors "github.com/rotisserie/eris"
)

// Call kubectl port-forward. Callers are expected to clean up the returned portFwd *exec.cmd after the port-forward is no longer needed.
func PortForward(namespace string, resource string, localPort string, kubePort string, verbose bool) (*exec.Cmd, error) {

	/** port-forward command **/

	portFwd := exec.Command("kubectl", "port-forward", "-n", namespace,
		resource, fmt.Sprintf("%s:%s", localPort, kubePort))

	portFwd.Stderr = os.Stderr
	if verbose {
		portFwd.Stdout = os.Stdout
	}

	if err := portFwd.Start(); err != nil {
		return nil, err
	}

	return portFwd, nil

}

// Call kubectl port-forward and make a GET request.
// Callers are expected to clean up the returned portFwd *exec.cmd after the port-forward is no longer needed.
func PortForwardGet(ctx context.Context, namespace string, resource string, localPort string, kubePort string, verbose bool, getPath string) (string, *exec.Cmd, error) {

	/** port-forward command **/

	portFwd, err := PortForward(namespace, resource, localPort, kubePort, verbose)
	if err != nil {
		return "", nil, err
	}

	localCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	// wait for port-forward to be ready
	retryInterval := time.Millisecond * 250
	result := make(chan string)
	errs := make(chan error)
	go func() {
		for {
			select {
			case <-localCtx.Done():
				return
			default:
			}
			res, err := http.Get("http://localhost:" + localPort + getPath)
			if err != nil {
				errs <- err
				time.Sleep(retryInterval)
				continue
			}
			if res.StatusCode != 200 {
				errs <- errors.Errorf("invalid status code: %v %v", res.StatusCode, res.Status)
				time.Sleep(retryInterval)
				continue
			}
			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				errs <- err
				time.Sleep(retryInterval)
				continue
			}
			res.Body.Close()
			result <- string(b)
			return
		}
	}()

	var multiErr *multierror.Error
	for {
		select {
		case err := <-errs:
			multiErr = multierror.Append(multiErr, err)
		case res := <-result:
			return res, portFwd, nil
		case <-localCtx.Done():
			if portFwd.Process != nil {
				portFwd.Process.Kill()
				portFwd.Process.Release()
			}
			return "", nil, errors.Errorf("timed out trying to connect to localhost during port-forward, errors: %v", multiErr)
		}
	}

}

func GetFreePort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	tcpAddr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errors.Errorf("Error occured looking for an open tcp port")
	}
	return tcpAddr.Port, nil
}
