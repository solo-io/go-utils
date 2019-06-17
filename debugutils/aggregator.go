package debugutils

import (
	"io"
	"os"
	"path/filepath"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/tarutils"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ZipStorageClient interface {
	Save(fileName string, data io.Reader) error
	Read(fileName string) (io.ReadCloser, error)
}

type LocalZipStorageClient struct {
	fs afero.Fs
}

func NewLocalZipStorageClient(fs afero.Fs) *LocalZipStorageClient {
	return &LocalZipStorageClient{fs: fs}
}

func (lsc *LocalZipStorageClient) Save(fileName string, data io.Reader) error {
	file, err := lsc.fs.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, data)
	if err != nil {
		return err
	}
	return nil
}

func (lsc *LocalZipStorageClient) Read(file string) (io.ReadCloser, error) {
	return lsc.fs.OpenFile(file, os.O_RDWR, 0777)
}

const (
	aggregatorName = "aggregator"
)

type Aggregator struct {
	resourceCollector ResourceCollector
	podFinder         PodFinder
	logCollector      *LogCollector
	zipStorageClient  ZipStorageClient
	fs                afero.Fs

	dir string
}

func NewAggregator(collector ResourceCollector, podFinder PodFinder, logCollector *LogCollector) *Aggregator {
	return &Aggregator{resourceCollector: collector, podFinder: podFinder, logCollector: logCollector}
}

func DefaultAggregator() (*Aggregator, error) {
	podFinder, err := NewLabelPodFinder()
	if err != nil {
		return nil, errors.InitializationError(err, aggregatorName)
	}
	collector, err := NewResourceCollector()
	if err != nil {
		return nil, errors.InitializationError(err, aggregatorName)
	}
	logCollector, err := DefaultLogCollector()
	if err != nil {
		return nil, errors.InitializationError(err, aggregatorName)
	}
	fs := afero.NewOsFs()
	tmpd, err := afero.TempDir(fs, "", "")
	if err != nil {
		return nil, err
	}
	storageClient := NewLocalZipStorageClient(fs)
	return &Aggregator{
		logCollector:      logCollector,
		podFinder:         podFinder,
		resourceCollector: collector,
		fs:                fs,
		dir:               tmpd,
		zipStorageClient:  storageClient,
	}, nil

}

func (a *Aggregator) StreamFromManifest(manifest helmchart.Manifests, namespace, filename string) error {
	if err := a.createSubResourceDirectories(); err != nil {
		return err
	}
	unstructuredResources, err := manifest.ResourceList()
	if err != nil {
		return err
	}
	kubeResources, err := a.resourceCollector.RetrieveResources(unstructuredResources, namespace, metav1.ListOptions{})
	if err != nil {
		return err
	}
	if err := a.resourceCollector.SaveResources(filepath.Join(a.dir, "resources"), kubeResources); err != nil {
		return err
	}
	logRequests, err := a.logCollector.GetLogRequests(unstructuredResources)
	if err != nil {
		return err
	}
	if err = a.logCollector.SaveLogs(filepath.Join(a.dir, "logs"), logRequests); err != nil {
		return err
	}
	tarball, err := afero.TempFile(a.fs, "", "")
	defer a.fs.Remove(tarball.Name())
	if err != nil {
		return err
	}
	if err := tarutils.Tar(a.dir, a.fs, tarball); err != nil {
		return err
	}
	_, err = tarball.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	if err := a.zipStorageClient.Save(filename, tarball); err != nil {
		return err
	}
	return nil
}

func (a *Aggregator) createSubResourceDirectories() error {
	directories := []string{"resources", "logs"}
	for _, v := range directories {
		resourceDir := filepath.Join(a.dir, v)
		err := a.fs.Mkdir(resourceDir, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}
