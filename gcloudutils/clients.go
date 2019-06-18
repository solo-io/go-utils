package gcloudutils

import (
	"context"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudbuild/v1"
)

func NewCloudBuildClient(ctx context.Context) (*cloudbuild.Service, error) {
	googleHttpClient, err := google.DefaultClient(ctx, cloudbuild.CloudPlatformScope)
	if err != nil {
		return nil, err
	}
	buildClient, err := cloudbuild.New(googleHttpClient)
	if err != nil {
		return nil, err
	}
	return buildClient, nil
}
