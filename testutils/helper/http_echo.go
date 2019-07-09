package helper

import (
	"time"
)

type echoPod struct {
	*testContainer
}

func (t *echoPod) Deploy(timeout time.Duration) error {
	return t.deploy(timeout)
}

const (
	defaultHttpEchoImage = "kennship/http-echo:latest"
	HttpEchoName         = "http-echo"
	HttpEchoPort         = 3000
)

func NewEchoHttp(namespace string) (*echoPod, error) {
	container, err := newTestContainer(namespace, defaultHttpEchoImage, HttpEchoName, HttpEchoPort)
	if err != nil {
		return nil, err
	}
	return &echoPod{
		testContainer: container,
	}, nil
}

const (
	defaultTcpEchoImage = "soloio/tcp-echo:latest"
	TcpEchoName         = "tcp-echo"
	TcpEchoPort         = 1025
)

func NewEchoTcp(namespace string) (*echoPod, error) {
	container, err := newTestContainer(namespace, defaultTcpEchoImage, TcpEchoName, TcpEchoPort)
	if err != nil {
		return nil, err
	}
	return &echoPod{
		testContainer: container,
	}, nil
}
