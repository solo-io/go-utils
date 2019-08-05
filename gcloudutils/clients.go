package gcloudutils

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/option"
)

const (
	gceDirEnv = "GOOGLE_APPLICATION_DIR"
)

var (
	DefaultRoot = "/etc/gce"
)

func init() {
	if gceDir := os.Getenv(gceDirEnv); gceDir != "" {
		DefaultRoot = gceDir
	}
}

func configFileName(projectId string) string {
	return filepath.Join(DefaultRoot, fmt.Sprintf("%s-creds.json", projectId))
}

func credsFromProjectId(ctx context.Context, projectId string) (*google.Credentials, error) {
	pathToCredsFile := configFileName(projectId)
	contextutils.LoggerFrom(ctx).Infow("Looking for creds for project",
		zap.String("projectId", projectId),
		zap.String("credsFile", pathToCredsFile))
	credByt, err := ioutil.ReadFile(pathToCredsFile)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error reading creds file",
			zap.Error(err),
			zap.String("projectId", projectId),
			zap.String("credsFile", pathToCredsFile))
		return nil, err
	}
	creds, err := google.CredentialsFromJSON(ctx, credByt, cloudbuild.CloudPlatformScope)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error getting creds",
			zap.Error(err),
			zap.String("projectId", projectId),
			zap.String("credsFile", pathToCredsFile))
		return nil, err
	}
	return creds, nil
}

func NewCloudBuildClient(ctx context.Context, projectId string) (*cloudbuild.Service, error) {
	creds, err := credsFromProjectId(ctx, projectId)
	if err != nil {
		return nil, err
	}
	googleHttpClient := oauth2.NewClient(ctx, creds.TokenSource)
	buildClient, err := cloudbuild.New(googleHttpClient)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error creating cloud build client",
			zap.Error(err),
			zap.String("proejctId", projectId))
		return nil, err
	}
	return buildClient, nil
}

func NewStorageClient(ctx context.Context, projectId string) (*storage.Client, error) {
	creds, err := credsFromProjectId(ctx, projectId)
	if err != nil {
		return nil, err
	}
	return storage.NewClient(ctx, option.WithTokenSource(creds.TokenSource))
}

func NewPubSubClient(ctx context.Context, projectId string) (*pubsub.Client, error) {
	creds, err := credsFromProjectId(ctx, projectId)
	if err != nil {
		return nil, err
	}
	pubsub, err := pubsub.NewClient(ctx, projectId, option.WithTokenSource(creds.TokenSource))
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error creating pub sub client",
			zap.Error(err),
			zap.String("proejctId", projectId))
		return nil, err
	}
	return pubsub, nil
}
