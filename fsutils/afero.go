package fsutils

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
)

// These two utils copied from solobot
func SetupTemporaryFiles(fs afero.Fs) (file afero.File, dir string, err error) {
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

func GetRepoFolder(fs afero.Fs, tmpd string) (string, error) {
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
