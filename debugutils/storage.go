package debugutils

import (
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

//go:generate mockgen -destination=./mocks/storage.go -source storage.go -package mocks

type StorageObject struct {
	resource io.Reader
	name string
}

type StorageClient interface {
	Save(location string, resources ...*StorageObject) error
}

type FileStorageClient struct {
	fs afero.Fs
}

func NewFileStorageClient(fs afero.Fs) *FileStorageClient {
	return &FileStorageClient{fs: fs}
}

func (fsc *FileStorageClient) Save(location string, resources ...*StorageObject) error {
	for _, resource := range resources {
		fileName := filepath.Join(location, resource.name)
		file, err := fsc.fs.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0777)
		if err != nil {
			return err
		}
		_, err = io.Copy(file, resource.resource)
		if err != nil {
			return err
		}
	}
	return nil
}

