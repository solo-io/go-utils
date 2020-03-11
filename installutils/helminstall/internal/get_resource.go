package internal

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rotisserie/eris"
)

//go:generate mockgen -destination mocks/mock_resource_fetcher.go -source ./get_resource.go

type ResourceFetcher interface {
	GetResource(uri string) (io.ReadCloser, error)
}

func NewDefaultResourceFetcher() ResourceFetcher {
	return &resourceFetcher{}
}

type resourceFetcher struct{}

// Get the resource identified by the given URI.
// The URI can either be an http(s) address or a relative/absolute file path.
func (r *resourceFetcher) GetResource(uri string) (io.ReadCloser, error) {
	var file io.ReadCloser
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		resp, err := http.Get(uri)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, eris.Errorf("http GET returned status %d for resource %s", resp.StatusCode, uri)
		}

		file = resp.Body
	} else {
		path, err := filepath.Abs(uri)
		if err != nil {
			return nil, eris.Wrapf(err, "getting absolute path for %v", uri)
		}

		f, err := os.Open(path)
		if err != nil {
			return nil, eris.Wrapf(err, "opening file %v", path)
		}
		file = f
	}

	// Write the body to file
	return file, nil
}
