package builders

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/gcloudutils"

	"cloud.google.com/go/storage"
	"github.com/ghodss/yaml"
	"github.com/google/go-github/v32/github"
	"github.com/rs/zerolog"
	"google.golang.org/api/cloudbuild/v1"
)

type StorageBuilder struct {
	client *storage.Client
}

func NewStorageBuilder(ctx context.Context, projectId string) (*StorageBuilder, error) {
	client, err := gcloudutils.NewStorageClient(ctx, projectId)
	if err != nil {
		return nil, err
	}
	return &StorageBuilder{client: client}, nil
}

func (sb *StorageBuilder) InitBuildWithSha(ctx context.Context, builderCtx ShaBuildContext) (*cloudbuild.Build, error) {
	cbm, content, err := unmarshalCloudbuild(ctx, builderCtx, builderCtx.Sha())
	if err != nil {
		return nil, err
	}

	attrs, err := sb.copyArchiveToStorage(ctx, builderCtx, builderCtx.Sha())
	if err != nil {
		return nil, err
	}

	cbm.Source = &cloudbuild.Source{
		StorageSource: &cloudbuild.StorageSource{
			Object: attrs.Name,
			Bucket: attrs.Bucket,
		},
	}

	if strings.Contains(content, ToEnv(COMMIT_SHA)) {
		cbm.Substitutions = map[string]string{
			COMMIT_SHA: builderCtx.Sha(),
		}
	}
	cbm.Substitutions = addDefaultSubsitutions(content, cbm)
	return cbm, nil
}

func (sb *StorageBuilder) InitBuildWithTag(ctx context.Context, builderCtx TagBuildContext) (*cloudbuild.Build, error) {
	cbm, content, err := unmarshalCloudbuild(ctx, builderCtx, builderCtx.Tag())
	if err != nil {
		return nil, err
	}
	attrs, err := sb.copyArchiveToStorage(ctx, builderCtx, builderCtx.Tag())
	if err != nil {
		return nil, err
	}

	cbm.Source = &cloudbuild.Source{
		StorageSource: &cloudbuild.StorageSource{
			Object: attrs.Name,
			Bucket: attrs.Bucket,
		},
	}

	cbm.Substitutions = make(map[string]string)

	if strings.Contains(content, ToEnv(TAG_NAME)) {
		cbm.Substitutions[TAG_NAME] = builderCtx.Tag()
	}
	if strings.Contains(content, ToEnv(COMMIT_SHA)) {
		cbm.Substitutions[COMMIT_SHA] = builderCtx.Sha()
	}

	cbm.Substitutions = addDefaultSubsitutions(content, cbm)
	return cbm, nil
}

func (sb *StorageBuilder) copyArchiveToStorage(ctx context.Context, builderCtx BuildContext, ref string) (*storage.ObjectAttrs, error) {
	logger := zerolog.Ctx(ctx)
	client, owner, repo := builderCtx.Client(), builderCtx.Owner(), builderCtx.Repo()
	opts := &github.RepositoryContentGetOptions{
		Ref: ref,
	}

	archiveURL, _, err := client.Repositories.GetArchiveLink(ctx, owner, repo, github.Tarball, opts, true)
	if err != nil {
		logger.Error().Err(err).Msg("can't get archive")
		return nil, err
	}

	tmpf, err := os.CreateTemp("", "*.tar.gz")
	if err != nil {
		logger.Error().Err(err).Msg("can't create temp file")
		return nil, err
	}
	tmpf.Close()
	defer os.Remove(tmpf.Name())

	err = downloadFile(archiveURL.String(), tmpf.Name())
	if err != nil {
		logger.Error().Err(err).Msg("can't download file")
		return nil, err
	}
	cleantmpf, err := removeGHPrefix(ctx, tmpf.Name())
	if err != nil {
		logger.Error().Err(err).Msg("can't clean temp file")
		return nil, err
	}
	defer os.Remove(cleantmpf.Name())

	obj, err := sb.copyToBucket(ctx, builderCtx, ref, cleantmpf.Name())
	if err != nil {
		return nil, err
	}

	attrs, err := obj.Attrs(ctx)

	return attrs, nil
}

func (sb *StorageBuilder) copyToBucket(ctx context.Context, builderCtx BuildContext, ref, archiveFile string) (*storage.ObjectHandle, error) {
	logger := zerolog.Ctx(ctx)

	reader, err := os.Open(archiveFile)
	defer reader.Close()
	if err != nil {
		logger.Error().Err(err).Msg("can't open archive file")
		return nil, err
	}

	bucket := bucketName(builderCtx.ProjectId())
	name := fmt.Sprintf(stagingLocation()+"/%s/%s_%d.tgz", builderCtx.Repo(), ref, time.Now().Unix())

	bh := sb.client.Bucket(bucket)
	// Next check if the bucket exists
	if _, err = bh.Attrs(ctx); err != nil {
		logger.Error().Err(err).Msg(fmt.Sprintf("Bucket %s doesn't exist", bucket))
		return nil, err
	}

	obj := bh.Object(name)
	w := obj.NewWriter(ctx)
	if _, err := io.Copy(w, reader); err != nil {
		return nil, errors.Wrapf(err, "could not write build zip to bucket")
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	return obj, err
}

func removeGHPrefix(ctx context.Context, archiveFile string) (*os.File, error) {
	logger := zerolog.Ctx(ctx)

	tmpf, err := os.CreateTemp("", "*.tar.gz")
	if err != nil {
		logger.Error().Err(err).Msg("can't create temp file")
		return nil, err
	}
	defer tmpf.Close()

	reader, err := os.Open(archiveFile)
	defer reader.Close()

	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	gzipWriter := gzip.NewWriter(tmpf)
	defer gzipWriter.Close()

	tarReader := tar.NewReader(gzipReader)
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		header.Name = removeGHPrefixFromPath(header.Name)
		tarWriter.WriteHeader(header)
		if header.Size != 0 {
			_, err = io.Copy(tarWriter, tarReader)
			if err != nil {
				return nil, err
			}
		}
	}

	return tmpf, nil
}

func stagingLocation() string {
	return "source"
}

func bucketName(proj string) string {
	return proj + "_cloudbuild"
}

func removeGHPrefixFromPath(path string) string {
	// we want to transform this
	//    solo-io-envoy-gloo-96ea44f/test/mocks/nats/streaming/BUILD
	// to this:
	//    test/mocks/nats/streaming/BUILD

	// so basically remove the first path of the path.
	i := strings.Index(path, "/")
	return path[i+1:]

}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func unmarshalCloudbuild(ctx context.Context, builderCtx BuildContext, ref string) (*cloudbuild.Build, string, error) {
	content, err := getCloudBuildYaml(ctx, builderCtx, ref)
	if err != nil {
		return nil, "", err
	}

	var cbm cloudbuild.Build
	if err := yaml.Unmarshal([]byte(content), &cbm); err != nil {
		return nil, "", fmt.Errorf("unable to wrangle cloudbuild.yaml into cloudbuild message, %s", err)
	}
	return &cbm, content, nil
}

func addDefaultSubsitutions(file string, build *cloudbuild.Build) map[string]string {
	var subs map[string]string
	if build.Substitutions == nil {
		subs = make(map[string]string)
	} else {
		subs = build.Substitutions
	}
	if strings.Contains(file, ToEnv(REPO_NAME)) && build.Source.RepoSource != nil {
		subs[REPO_NAME] = build.Source.RepoSource.RepoName
	}
	return subs
}

func getCloudBuildYaml(ctx context.Context, builderCtx BuildContext, location string) (string, error) {
	owner, repo, client := builderCtx.Owner(), builderCtx.Repo(), builderCtx.Client()

	opt := &github.RepositoryContentGetOptions{
		Ref: location,
	}

	cloudbuildFile, _, _, err := client.Repositories.GetContents(ctx, owner, repo, gcloudutils.CloudbuildFile, opt)
	if err != nil {
		return "", err
	}

	return cloudbuildFile.GetContent()
}
