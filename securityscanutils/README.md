## Trivy Security Scanning

Trivy is a security scanning tool which we use to scan our images for vulnerabilities.
You can run a trivy scan identical to CI on your own command line by installing trivy and running
```shell
trivy image --severity HIGH,CRITICAL quay.io/solo-io/<IMAGE>:<VERSION>
```

## Using securityscanutils
The following code snippet shows how to import and use the `SecurityScanner` to scan a repositories' releases. Multiple
repositories can be specified for scanning. 

The `GITHUB_TOKEN` environment variable must be set for security scanning to work.

```go
package main

import (
	"context"
	"log"

	"github.com/Masterminds/semver/v3"
	. "github.com/solo-io/go-utils/securityscanutils"
)

func main() {
    // This is a constraint on which releases from the repository are scanned.
    // Any releases that don't pass this constraint will not be scanned. Passed into the `VersionConstraint` option.
	constraint, _ := semver.NewConstraint(">= v1.7.0")
	scanner := SecurityScanner{
		Repos: []*SecurityScanRepo{
			{
				Repo:  "gloo",
				Owner: "solo-io",
				Opts: &SecurityScanOpts{
					OutputDir: "_output/scans",
                    // Different release versions may have different images to scan.
                    // In this example, we introduced the "discovery" image in 1.7.0, and
                    // specify the constraint as such. 
                    // Each version should only match only ONE constraint, else an error will be thrown.
                    // Read https://github.com/Masterminds/semver#checking-version-constraints for more about how to use
                    // semver constraints
					ImagesPerVersion: map[string][]string{
					    "1.7.x": {"gloo", "gloo-envoy-wrapper"},
						">=v1.7.0 <= v1.8.0": {"gloo", "gloo-envoy-wrapper", "discovery"},
					},
                    // If VersionConstraint is not specified, all releases from the repo will be scanned, including
                    // pre-releases, which is not recommended.
					VersionConstraint:      constraint,
					ImageRepo:              "quay.io/solo-io",
                    // Setting this to true will upload any generated sarif files to the github repository
                    // endpoint, e.g. https://github.com/solo-io/gloo/security/code-scanning
                    // read more here: https://docs.github.com/en/rest/reference/code-scanning
					UploadCodeScanToGithub: true,
					// Opens/Updates Github Issue for each version that has images that have vulnerabilities
                    CreateGithubIssuePerVersion: true,
				},
			},
		},
	}
	err := scanner.GenerateSecurityScans(context.Background())
	if err != nil {
		log.Fatalf(err.Error())
	}
}
```