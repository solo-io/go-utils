package installutils

import (
	"path/filepath"

	"github.com/solo-io/go-utils/vfsutils"
	"github.com/spf13/afero"
	"k8s.io/helm/pkg/manifest"
)

func GetManifestsFromRemoteTar(tarUrl string) ([]manifest.Manifest, error) {
	fs := afero.NewMemMapFs()
	dir, err := vfsutils.MountTar(fs, tarUrl)
	if err != nil {
		return nil, err
	}
	files, err := afero.ReadDir(fs, dir)
	if err != nil {
		return nil, err
	}
	templates := make(map[string]string)
	for _, file := range files {
		filename := filepath.Join(dir, file.Name())
		contents, err := afero.ReadFile(fs, filename)
		if err != nil {
			return nil, err
		}
		templates[filename] = string(contents)
	}
	return manifest.SplitManifests(templates), nil
}
