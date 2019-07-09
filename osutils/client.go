package osutils

import (
	"io/ioutil"
	"os"
)

type OsClient interface {
	Getenv(key string) string
	ReadFile(path string) ([]byte, error)
}

type osClient struct {
}

func (*osClient) Getenv(key string) string {
	return os.Getenv(key)
}

func (*osClient) ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func NewOsClient() OsClient {
	return &osClient{}
}
