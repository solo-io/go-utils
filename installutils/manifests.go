package installutils

import (
	"path/filepath"
	"regexp"
	"strings"

	"helm.sh/helm/v3/pkg/releaseutil"

	"github.com/solo-io/go-utils/vfsutils"
	"github.com/spf13/afero"
)

var (
	kindRegex = regexp.MustCompile("kind:(.*)\n")
)

func GetManifestsFromRemoteTar(tarUrl string) ([]releaseutil.Manifest, error) {
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
	return SplitManifests(templates), nil
}

// SplitManifests takes a map of rendered templates and splits them into the
// detected manifests.
// (ported from Helm 2: https://github.com/helm/helm/blob/release-2.16/pkg/manifest/splitter.go)
func SplitManifests(templates map[string]string) []releaseutil.Manifest {
	var listManifests []releaseutil.Manifest
	// extract kind and name
	for k, v := range templates {
		match := kindRegex.FindStringSubmatch(v)
		h := "Unknown"
		if len(match) == 2 {
			h = strings.TrimSpace(match[1])
		}
		m := releaseutil.Manifest{Name: k, Content: v, Head: &releaseutil.SimpleHead{Kind: h}}
		listManifests = append(listManifests, m)
	}

	return listManifests
}
