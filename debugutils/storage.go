package debugutils

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/spf13/afero"
	"golang.org/x/sync/errgroup"
)

//go:generate mockgen -destination=./mocks/storage.go -source storage.go -package mocks

type StorageObject struct {
	resource io.Reader
	name     string
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

type GcsStorageClient struct {
	client *storage.Client
	ctx context.Context
}

func NewGcsStorageClient(ctx context.Context) (*GcsStorageClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &GcsStorageClient{client: client, ctx: ctx}, nil
}


func (gsc *GcsStorageClient) Save(location string, resources ...*StorageObject) error {
	bucket := gsc.client.Bucket("location")
	eg := errgroup.Group{}
	for _, resource := range resources {
		resource := resource
		eg.Go(func() error {
			obj := bucket.Object(resource.name)
			w := obj.NewWriter(gsc.ctx)
			_, err := io.Copy(w, resource.resource)
			if err != nil {
				return err
			}
			return nil
		})
	}
	return eg.Wait()
}
