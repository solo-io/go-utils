package helper

import (
	"time"
)

const (
	defaultHttpEchoImage = "kennship/http-echo:latest"
	HttpEchoName         = "http-echo"
	HttpEchoPort         = 3000

)


func NewEchoHttp(namespace string) (*HttpEcho, error) {
	container, err := newTestContainer(namespace, defaultHttpEchoImage, HttpEchoName, HttpEchoPort)
	if err != nil {
		return nil, err
	}
	return &HttpEcho{
		TestContainer: container,
	}, nil
}

type HttpEcho struct {
	*TestContainer
}

func (t *HttpEcho) Deploy(timeout time.Duration) error {
	return t.deploy(timeout)
}
