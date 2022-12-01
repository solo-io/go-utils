## Trivy Security Scanning

Trivy is a security scanning tool which we use to scan our images for vulnerabilities.
You can run a trivy scan identical to CI on your own command line by installing trivy and running
```shell
trivy image --severity HIGH,CRITICAL quay.io/solo-io/<IMAGE>:<VERSION>
```

## Using securityscanutils
Using the utils here is as easy as using the CLI defined in the cli subdirectory. The snippet
below shows the output the said CLI's `help` command.

The `GITHUB_TOKEN` environment variable must be set for security scanning to work.

```bash
go-utils/securityscan % go run ./cli/main.go help

CLI for identifying CVEs in images

Usage:
  cvectl [command]

Available Commands:
  format-results Pull down security scan files from gcloud bucket and generate docs markdown file
  help           Help about any command
  scan-repo      Run Trivy scans against images for the repo specified and upload scan results to a google cloud bucket
  scan-version   Run Trivy scans against images for a single version

Flags:
  -h, --help      help for cvectl
  -v, --verbose   Enable verbose logging

Use "cvectl [command] --help" for more information about a command.
```