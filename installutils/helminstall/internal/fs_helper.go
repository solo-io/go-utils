package internal

import (
	"os"

	"github.com/spf13/afero"
)

//go:generate mockgen -source ./fs_helper.go -destination ./mocks/mock_fs_helper.go

// interface around needed afero functions
type FsHelper interface {
	NewTempFile(dir, prefix string) (f afero.File, err error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
	RemoveAll(path string) error
}

type tempFile struct {
	fs afero.Fs
}

func NewFs(fs afero.Fs) FsHelper {
	return &tempFile{fs: fs}
}

func (t *tempFile) NewTempFile(dir, prefix string) (f afero.File, err error) {
	return afero.TempFile(t.fs, dir, prefix)
}

func (t *tempFile) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return afero.WriteFile(t.fs, filename, data, perm)
}

func (t *tempFile) RemoveAll(path string) error {
	return t.fs.RemoveAll(path)
}
