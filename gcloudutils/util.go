package gcloudutils

import (
	"context"
	"fmt"

	"google.golang.org/api/cloudbuild/v1"
)

const (
	nameConst    = "name"
	versionConst = "version"
	constraint   = "constraint"
)

func HandleFailedSourceBuild(ctx context.Context, build *cloudbuild.Build) error {
	if build.StatusDetail != "" {
		if IsMissingSourceError(build.StatusDetail) {
			return fmt.Errorf("unable to resolve source for build %s: skipping handle", build.Id)
		}
	}
	return nil
}
