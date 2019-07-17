package gcloudutils

import (
	"context"

	"google.golang.org/api/cloudbuild/v1"
)

type CloudBuildRegistry struct {
	eventHandlers []CloudBuildEventHandler
}

func (r *CloudBuildRegistry) AddEventHandler(handler CloudBuildEventHandler) {
	r.eventHandlers = append(r.eventHandlers, handler)
}

type CloudBuildEventHandler interface {
	CloudBuild(ctx context.Context, build *cloudbuild.Build) error
}
