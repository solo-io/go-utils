package githubutils

import (
	"context"
	"github.com/google/go-github/github"
	"github.com/solo-io/go-utils/versionutils"
	"log"
	"os"
)

type UploadReleaseAssetSpec struct {
	Owner  string
	Repo   string
	Assets map[string]*os.File
}

func UploadReleaseAssetCli(spec *UploadReleaseAssetSpec) {
	version := versionutils.GetReleaseVersionOrExitGracefully()
	ctx := context.TODO()
	client := GetClientOrExit(ctx)
	release := GetReleaseOrExit(ctx, client, version, spec)
	UploadReleaseAssetsOrExit(ctx, client, release, spec)
}

func UploadReleaseAssetsOrExit(ctx context.Context, client *github.Client, release *github.RepositoryRelease, spec *UploadReleaseAssetSpec) {
	for name, asset := range spec.Assets {
		opts := &github.UploadOptions{
			Name: name,
		}
		_, _, err := client.Repositories.UploadReleaseAsset(ctx, spec.Owner, spec.Repo, release.GetID(), opts, asset)
		if err != nil {
			log.Fatalf("Error uploading assets. Error was: %s", err.Error())
		}
	}

}

func GetClientOrExit(ctx context.Context) *github.Client {
	client, err := GetClient(ctx)
	if err != nil {
		log.Fatalf("Could not get github client. Error was: %s", err.Error())
	}
	return client
}

func GetReleaseOrExit(ctx context.Context, client *github.Client, version *versionutils.Version, spec *UploadReleaseAssetSpec) *github.RepositoryRelease {
	release, _, err := client.Repositories.GetReleaseByTag(ctx, spec.Owner, spec.Repo, version.String())
	if err != nil {
		log.Fatalf("Could not find release %s. Error was: %s", version.String(), err.Error())
	}
	return release
}
