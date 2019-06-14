package debugutils

import (
	"io"
	"os"

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
	col ResourceCollector
	pf  PodFinder
	lsc *LogStorageClient
	zsc ZipStorageClient
	fs  afero.Fs

	dir string
	lrb *LogRequestBuilder
}

func NewAggregator(collector ResourceCollector, podFinder PodFinder, logStorage *LogStorageClient) *Aggregator {
	return &Aggregator{col: collector, pf: podFinder, lsc: logStorage}
}

func NewDefaultAggregator() (*Aggregator, error) {
	podFinder, err := NewLabelPodFinder()
	if err != nil {
		return nil, initializationError(err, aggregatorName)
	}
	collector, err := NewResourceCollector()
	if err != nil {
		return nil, initializationError(err, aggregatorName)
	}
	lrb, err := NewLogRequestBuilder()
	if err != nil {
		return nil, initializationError(err, aggregatorName)
	}
	fs := afero.NewOsFs()
	tmpd, err := afero.TempDir(fs, "", "")
	if err != nil {
		return nil, err
	}
	logStorage := NewLogFileStorage(fs, tmpd)
	storageClient := NewLocalZipStorageClient(fs)
	return &Aggregator{
		pf:  podFinder,
		col: collector,
		fs:  fs,
		dir: tmpd,
		lsc: logStorage,
		lrb: lrb,
		zsc: storageClient,
	}, nil

}

func (a *Aggregator) StreamFromManifest(manifest helmchart.Manifests, namespace, filename string) error {
	unstructuredResources, err := manifest.ResourceList()
	if err != nil {
		return err
	}
	kubeResources, err := a.col.RetrieveResources(unstructuredResources, namespace, metav1.ListOptions{})
	if err != nil {
		return err
	}
	if err := a.col.SaveResources(kubeResources, a.fs, a.dir); err != nil {
		return err
	}
	logRequests, err := a.lrb.LogsFromUnstructured(unstructuredResources)
	if err != nil {
		return err
	}
	if err = a.lsc.FetchLogs(logRequests); err != nil {
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
	if err := a.zsc.Save(filename, tarball); err != nil {
		return err
	}
	return nil
}
