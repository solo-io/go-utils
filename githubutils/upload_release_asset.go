package githubutils

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/avast/retry-go"
	"github.com/google/go-github/github"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/versionutils"
)

type ReleaseAssetSpec struct {
	Name       string
	ParentPath string
	UploadSHA  bool
}

type UploadReleaseAssetSpec struct {
	Owner             string
	Repo              string
	Assets            []ReleaseAssetSpec
	SkipAlreadyExists bool
}

func UploadReleaseAssetCli(spec *UploadReleaseAssetSpec) {
	version := versionutils.GetReleaseVersionOrExitGracefully()
	ctx := context.TODO()
	client := GetClientOrExit(ctx)
	release := GetReleaseOrExit(ctx, client, version, spec)
	uploadReleaseAssetsOrExit(ctx, client, release, spec)
}

func uploadReleaseAssetsOrExit(ctx context.Context, client *github.Client, release *github.RepositoryRelease, spec *UploadReleaseAssetSpec) {
	for _, assetSpec := range spec.Assets {
		uploadReleaseAssetOrExit(ctx, client, release, spec, assetSpec)
	}
}

func uploadReleaseAssetOrExit(ctx context.Context, client *github.Client, release *github.RepositoryRelease, spec *UploadReleaseAssetSpec, asset ReleaseAssetSpec) {
	if spec.SkipAlreadyExists && assetAlreadyExists(release, asset.Name) {
		return
	}
	path := filepath.Join(asset.ParentPath, asset.Name)
	uploadFileOrExit(ctx, client, release, spec, asset.Name, path)
	if asset.UploadSHA {
		uploadShaOrExit(ctx, client, release, spec, asset)
	}
}

func uploadShaOrExit(ctx context.Context, client *github.Client, release *github.RepositoryRelease, spec *UploadReleaseAssetSpec, asset ReleaseAssetSpec) {
	path := filepath.Join(asset.ParentPath, asset.Name)
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	shaName := asset.Name + ".sha256"
	shaPath := filepath.Join(asset.ParentPath, shaName)
	writeSha256OrExit(ctx, file, shaPath)
	uploadFileOrExit(ctx, client, release, spec, shaName, shaPath)
}

func uploadFileOrExit(ctx context.Context, client *github.Client, release *github.RepositoryRelease, spec *UploadReleaseAssetSpec, name, path string) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalf("Error reading file %s: %s", path, err.Error())
	}

	// Using default retry settings for now, 10 attempts, 100ms delay with backoff
	err = retry.Do(func() error {
		return tryUploadAsset(ctx, client, release, spec, name, file)
	})

	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalf("Error uploading assets. Error was: %s", err.Error())
	}
}

func tryUploadAsset(ctx context.Context, client *github.Client, release *github.RepositoryRelease, spec *UploadReleaseAssetSpec, name string, file *os.File) error {
	opts := &github.UploadOptions{
		Name: name,
	}
	_, _, err := client.Repositories.UploadReleaseAsset(ctx, spec.Owner, spec.Repo, release.GetID(), opts, file)
	if err != nil {
		loadedRelease, _, _ := client.Repositories.GetRelease(ctx, spec.Owner, spec.Repo, release.GetID())
		if loadedRelease != nil {
			tryDeleteAsset(ctx, client, loadedRelease, spec, name)
		}
	}
	return err
}

func tryDeleteAsset(ctx context.Context, client *github.Client, release *github.RepositoryRelease, spec *UploadReleaseAssetSpec, name string) error {
	for _, asset := range release.Assets {
		if asset.GetName() == name {
			_, err := client.Repositories.DeleteReleaseAsset(ctx, spec.Owner, spec.Repo, asset.GetID())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func writeSha256OrExit(ctx context.Context, file *os.File, outputPath string) {
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		contextutils.LoggerFrom(ctx).Fatal(err)
	}
	sha256String := hex.EncodeToString(h.Sum(nil)) + " " + filepath.Base(file.Name()) + "\n"
	err := ioutil.WriteFile(outputPath, []byte(sha256String), 0700)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatal(err)
	}
}

func assetAlreadyExists(release *github.RepositoryRelease, name string) bool {
	for _, asset := range release.Assets {
		if asset.GetName() == name {
			return true
		}
	}
	return false
}

func GetClientOrExit(ctx context.Context) *github.Client {
	client, err := GetClient(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalf("Could not get github client. Error was: %s", err.Error())
	}
	return client
}

func GetReleaseOrExit(ctx context.Context, client *github.Client, version *versionutils.Version, spec *UploadReleaseAssetSpec) *github.RepositoryRelease {
	release, _, err := client.Repositories.GetReleaseByTag(ctx, spec.Owner, spec.Repo, version.String())
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalf("Could not find release %s. Error was: %s", version.String(), err.Error())
	}
	return release
}
