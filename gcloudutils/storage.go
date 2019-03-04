package main

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"io"
	"os"

	"cloud.google.com/go/storage"
)

const (
	distBucket         = "gloo-ee-distribution"
	distBucketReleases = "releases"
	distBucketTags     = "tags"
	indexFile          = "index.yaml"

	objectDNE = "storage: object doesn't exist"
)

var (
	projectId = os.Getenv("PROJECT_ID")
)

type GcloudStorageClient struct {
	ctx    context.Context
	client *storage.Client
}

func NewGcloudBucketClient(ctx context.Context) (*GcloudStorageClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	storageCli := &GcloudStorageClient{
		client: client,
		ctx:    ctx,
	}
	return storageCli, nil
}

func (db *GcloudStorageClient) CreateBucket(bucketName string) error {
	bucket := db.client.Bucket(bucketName)
	return bucket.Create(db.ctx, projectId, nil)
}

func (db *GcloudStorageClient) SaveFileToBucket(inputPath, bucketName, destPath string, entity storage.ACLEntity, role storage.ACLRole) error {
	logger := contextutils.LoggerFrom(db.ctx)
	bucket := db.client.Bucket(bucketName)
	obj := bucket.Object(destPath)
	logger.Infof("writing object: %v", obj.ObjectName())
	wr := obj.NewWriter(db.ctx)
	r, err := os.Open(inputPath)
	defer r.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(wr, r)
	if err != nil {
		return err
	}
	err = wr.Close()
	if err != nil {
		return err
	}
	if entity != "" && role != "" {
		err = obj.ACL().Set(db.ctx, storage.AllUsers, storage.RoleReader)
		if err != nil {
			return err
		}
	}
	return nil
}

