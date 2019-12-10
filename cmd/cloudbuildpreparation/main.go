package main

import (
	"context"
	"io/ioutil"

	"github.com/solo-io/go-utils/tarutils"
	"github.com/spf13/afero"

	"github.com/solo-io/go-utils/githubutils"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	err := run(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("unable to complete cloud build preparation", zap.Error(err))
	}

}

func run(ctx context.Context) error {

	// This utility expects GITHUB_TOKEN to exist in the environment
	githubClient, err := githubutils.GetClient(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Warnw("could not get github client", zap.Error(err))
		return err
	}

	// TODO pass as flags
	owner := "solo-io"
	repo := "gloo"
	sha := "master"
	outputDir := ""

	fs := afero.NewOsFs()
	//err = fs.MkdirAll("tmp", 0755)
	//if err != nil {
	//	return err
	//}
	file, err := ioutil.TempFile("", "new-file")
	if err != nil {
		return err
	}

	if err := githubutils.DownloadRepoArchive(ctx, githubClient, file, owner, repo, sha); err != nil {
		contextutils.LoggerFrom(ctx).Warnw("could not download repo", zap.Error(err))
		return err
	}

	err = tarutils.Untar(outputDir, file.Name(), fs)
	if err != nil {
		return err
	}
	return nil
}
