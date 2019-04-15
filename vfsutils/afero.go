package vfsutils

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/google/go-github/github"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/tarutils"

	"github.com/spf13/afero"
)

func MountCode(fs afero.Fs, ctx context.Context, client *github.Client, owner, repo, ref string) (dir string, err error) {
	tarFile, codeDir, err := setupTemporaryFiles(fs)
	if err != nil {
		return "", err
	}
	defer fs.Remove(tarFile.Name())
	if err := githubutils.DownloadRepoArchive(ctx, client, tarFile, owner, repo, ref); err != nil {
		return "", err
	}

	if err := tarutils.Untar(codeDir, tarFile.Name(), fs); err != nil {
		return "", err
	}

	repoFolderName, err := getRepoFolder(fs, codeDir)
	if err != nil {
		return "", err
	}
	return repoFolderName, nil
}

func setupTemporaryFiles(fs afero.Fs) (file afero.File, dir string, err error) {
	tmpf, err := afero.TempFile(fs, "", "tar-file-")
	if err != nil {
		return nil, "", err
	}

	tmpd, err := afero.TempDir(fs, "", "tar-dir-")
	if err != nil {
		return nil, "", err
	}
	return tmpf, tmpd, err
}

func getRepoFolder(fs afero.Fs, tmpd string) (string, error) {
	files, err := afero.ReadDir(fs, tmpd)
	if err != nil {
		return "", err
	}
	if len(files) != 1 {
		return "", fmt.Errorf("expected only one folder from archive tar, found (%d)", len(files))
	}

	var repoDirName string
	for _, file := range files {
		if file.IsDir() {
			repoDirName = file.Name()
		}
	}
	if repoDirName == "" {
		return "", fmt.Errorf("unable to find directory in archive of git repo")
	}
	return filepath.Join(tmpd, repoDirName), nil
}
