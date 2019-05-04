package helper

import (
	"time"
)

const (
	defaultHttpEchoImage = "kennship/http-echo:latest"
	HttpEchoName         = "http-echo"
	HttpEchoPort         = 3000

)


func NewEchoHttp(namespace string) (*httpEcho, error) {
	container, err := newTestContainer(namespace, defaultHttpEchoImage, HttpEchoName, HttpEchoPort)
	if err != nil {
		return nil, err
	}
	return &httpEcho{
		testContainer: container,
	}, nil
}

type httpEcho struct {
	*testContainer
}

func (t *httpEcho) Deploy(timeout time.Duration) error {
	return t.deploy(timeout)
}
