package gcloudutils

import (
	"context"
)

func IsPublic(ctx context.Context, projectId, repoName string) bool {
	// TODO: should this be in the app config?
	if projectId == "solo-public" {
		return true
	}
	// Support our testing efforts :) can be removed once
	// we figure out a better plan for live bot canary
	if repoName == "testrepo" {
		return true
	}
	return false
}
