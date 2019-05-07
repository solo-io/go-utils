package tarutils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

func Tar(src string, fs afero.Fs, writers ...io.Writer) error {
	srcInfo, err := fs.Stat(src)
	if err != nil {
		return fmt.Errorf("Unable to tar files - %v", err.Error())
	}

	var relativeRoot string
	if srcInfo.IsDir() {
		relativeRoot = srcInfo.Name()
	}

	mw := io.MultiWriter(writers...)

	gzw := gzip.NewWriter(mw)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()
	// walk path
	return afero.Walk(fs, src, func(file string, fi os.FileInfo, err error) error {
		var relativePath string
		if relativeRoot != "" {
			splitPath := strings.Split(filepath.Dir(file),  fmt.Sprintf("%s/", relativeRoot))
			if len(splitPath) == 2 {
				relativePath = splitPath[1]
			}
		}
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		f, err := fs.Open(file)
		if err != nil {
			return err
		}
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		f.Close()
		return nil
	})
}

func Untar(dst, src string, fs afero.Fs) error {
	file, err := fs.Open(src)
	if err != nil {
		return err
	}
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {
		// if no more files are found return
		case err == io.EOF:
			return nil
		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		target := filepath.Join(dst, header.Name)
		switch header.Typeflag {
		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := fs.Stat(target); err != nil {
				if err := fs.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		// if it's a file create it
		case tar.TypeReg:
			f, err := fs.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
			f.Close()
		}
	}
}
