
Helper for preparing Cloud Build environments


This binary has been designed to simplify Cloud Build initialization.
This should be run in a container as the first cloud build step.


## Features

- Pulls GitHub repo using GitHub client (and `GITHUB_TOKEN`)
  - Previously we had maintained a cache of the source code in a storage bucket.



## Building

1. This process is built in a self-contained multi-stage build
1. The process can then be copied into other containers


## Usage

### In Cloudbuild
- wip

### In Kubernetes
- wip

### With Docker run
- if you want the pulled source code to be placed in your host filesystem, make sure to use the same root path as your mounted volume.
  - for the below example your spec should include `outputDir: /mnt/mydata/optional-subdir`
```
docker run \
  -e GITHUB_TOKEN=`cat ~/.github/token` \
  -v $(pwd):/mnt/mydata \
  gcr.io/solo-io/cloudbuildpreparation:0.0.1 \
  --spec /mnt/mydata/sample-spec.yaml
```

### From inside the container (for testing)
- change the entrypoint to `sh` so you have a shell to work with
- provide your GitHub credential
- mount a volume to access the sample spec
```
docker run -it --entrypoint="/bin/sh" \
  -e GITHUB_TOKEN=`cat ~/.github/token` \
  -v $(pwd):/mnt/mydata \
  gcr.io/solo-io/cloudbuildpreparation:0.0.1
```
- verify that the sample file is available in the mounted directory
- execute the app
- verify that the source code was pulled to the specified directory
```
cat /mnt/mydata/sample-spec.yaml
./goapp --spec /mnt/mydata/sample-spec.yaml
ls temp-demo
# as desired, edit sample-spec.yaml and rerun
```

### Flags variant

```
go run main.go \
  --repo-name gloo \
  --repo-owner solo-io \
  --repo-sha master \
  --repo-output-dir test-dir
```
