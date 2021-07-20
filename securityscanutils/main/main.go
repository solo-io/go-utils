package main

import (
	"context"
	"log"

	"github.com/Masterminds/semver/v3"
	. "github.com/solo-io/go-utils/securityscanutils"
)

func main() {
	constraint, _ := semver.NewConstraint(">= v1.7.0")
	scanner := SecurityScanner{
		Repos: []*SecurityScanRepo{
				{
					Repo:  "gloo",
					Owner: "solo-io",
					Opts: &SecurityScanOpts{
						OutputDir: "_output/scans",
						ImagesPerVersion: map[string][]string{
							">=v1.7.0": {"gloo"},
						},
						VersionConstraint:      constraint,
						ImageRepo:              "quay.io/solo-io",
						UploadCodeScanToGithub: true,
					},
				},
		},
	}
	err := scanner.GenerateSecurityScans(context.Background())
	if err != nil {
		log.Fatalf(err.Error())
	}
}
