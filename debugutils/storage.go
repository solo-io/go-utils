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

type StorageObject struct {
	Resource io.Reader
	Name     string
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

func DefaultFileStorageClient() *FileStorageClient {
	return &FileStorageClient{fs: afero.NewOsFs()}
}

func (fsc *FileStorageClient) Save(location string, resources ...*StorageObject) error {
	for _, resource := range resources {
		fileName := filepath.Join(location, resource.Name)
		file, err := fsc.fs.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0777)
		if err != nil {
			return err
		}
		_, err = io.Copy(file, resource.Resource)
		if err != nil {
			return err
		}
		file.Close()
	}
	return nil
}

type GcsStorageClient struct {
	client *storage.Client
	ctx    context.Context
}

func NewGcsStorageClient(client *storage.Client, ctx context.Context) *GcsStorageClient {
	return &GcsStorageClient{client: client, ctx: ctx}
}

func DefaultGcsStorageClient(ctx context.Context) (*GcsStorageClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &GcsStorageClient{client: client, ctx: ctx}, nil
}

func (gsc *GcsStorageClient) Save(location string, resources ...*StorageObject) error {
	bucket := gsc.client.Bucket(location)
	eg := errgroup.Group{}
	for _, resource := range resources {
		resource := resource
		eg.Go(func() error {
			obj := bucket.Object(resource.Name)
			w := obj.NewWriter(gsc.ctx)
			defer w.Close()
			_, err := io.Copy(w, resource.Resource)
			if err != nil {
				return err
			}
			return nil
		})
	}
	return eg.Wait()
}
