package gcloudutils

import (
	"context"

	"github.com/google/go-github/github"

	"google.golang.org/api/cloudbuild/v1"
)

type CloudBuildRegistry struct {
	eventHandlers []CloudBuildEventHandler
}

func (r *CloudBuildRegistry) AddEventHandler(handler CloudBuildEventHandler) {
	r.eventHandlers = append(r.eventHandlers, handler)
}

type CloudBuildEventHandler interface {
	CloudBuild(ctx context.Context, client *github.Client, build *cloudbuild.Build) error
}
