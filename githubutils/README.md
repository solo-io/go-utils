## Uploading Release Assets to Github

To upload release assets to Github, follow these steps (requires go-utils 0.2.10+). 

### Create a Go script

Create a script `upload_github_release_asset.go`, like this: 

```go
package main

import (
	"github.com/solo-io/go-utils/githubutils"
)

func main() {
	assets := make([]githubutils.ReleaseAssetSpec, 2)
	assets[0] = githubutils.ReleaseAssetSpec{
		Name: "hello",
		ParentPath: "_output",
		UploadSHA: true,
	}
	assets[1] = githubutils.ReleaseAssetSpec{
		Name: "my-resource.yaml",
		ParentPath: "namespace",
	}
	spec := githubutils.UploadReleaseAssetSpec{
		Owner: "solo-io",
		Repo: "testrepo",
		Assets: assets,
		SkipAlreadyExists: true,
	}
	githubutils.UploadReleaseAssetCli(&spec)
}
```

### Create a Make target

```bash
#----------------------------------------------------------------------------------
# Github Assets
#----------------------------------------------------------------------------------

.PHONY: upload-github-release-assets
upload-github-release-assets: hello
	go run upload_github_release_assets.go
```

### Update cloudbuild.yaml to call this target

```yaml
steps:
- name: 'gcr.io/solo-corp/go-mod-make:0.1.1'
  args: [..., 'upload-github-release-assets', ...]
  secretEnv: ['GITHUB_TOKEN']
  env:
  - 'TAGGED_VERSION=$TAG_NAME'
```

Make sure `GITHUB_TOKEN` and `TAGGED_VERSION` are in the environment. 

### Notes

* On each asset, a flag `UploadSHA` can be set to true to upload a SHA256 hash file. 
* Set `SkipAlreadyExists=true` to not fail when trying to upload an asset that already exists. 