package tarutils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/solo-io/go-utils/errors"
	"github.com/spf13/afero"
)

func Tar(src string, fs afero.Fs, writers ...io.Writer) error {
	srcInfo, err := fs.Stat(src)
	if err != nil {
		return fmt.Errorf("Unable to tar files - %v", err.Error())
	}

	mw := io.MultiWriter(writers...)

	gzw := gzip.NewWriter(mw)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()
	// walk path
	return afero.Walk(fs, src, func(file string, fi os.FileInfo, err error) error {
		if file == src {
			return nil
		}

		filePrefix := getFilePrefix(srcInfo.IsDir(), src, file)
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		header.Name = filepath.Join(filePrefix, header.Name)

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

func getFilePrefix(srcDir bool, absoluteRoot, file string) string {
	var relativePath string
	if srcDir && file != absoluteRoot {
		pathToSrc := filepath.Dir(file)
		relativePath = strings.Replace(pathToSrc, absoluteRoot, "", 1)
	}
	return relativePath
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

// Retrieves the file from the given uri
// The uri can be a remote url or a file on the local fs
// This function returns a ReadCloser and it is the responsibility of the caller to close it
func RetrieveArchive(fs afero.Fs,  uri string) (io.ReadCloser, error) {
	var file io.ReadCloser
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		resp, err := http.Get(uri)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, errors.Errorf("http GET returned status %d", resp.StatusCode)
		}

		file = resp.Body
	} else {
		path, err := filepath.Abs(uri)
		if err != nil {
			return nil, errors.Wrapf(err, "getting absolute path for %v", uri)
		}

		f, err := fs.Open(path)
		if err != nil {
			return nil, errors.Wrapf(err, "opening file %v", path)
		}
		file = f
	}
	return file, nil
}
