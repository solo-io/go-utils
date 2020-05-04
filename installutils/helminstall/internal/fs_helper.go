package internal

import (
	"os"

	"github.com/solo-io/go-utils/installutils/helminstall/types"
	"github.com/spf13/afero"
)

type tempFile struct {
	fs afero.Fs
}

func NewFs(fs afero.Fs) types.FsHelper {
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
