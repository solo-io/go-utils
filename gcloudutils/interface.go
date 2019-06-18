package gcloudutils

import (
	"context"

	"github.com/google/go-github/github"

	"google.golang.org/api/cloudbuild/v1"
)

type Registry struct {
	eventHandlers []EventHandler
}

func (r *Registry) AddEventHandler(handler EventHandler) {
	r.eventHandlers = append(r.eventHandlers, handler)
}

type EventHandler interface {
	CloudBuild(ctx context.Context, client *github.Client, build *cloudbuild.Build) error
}
