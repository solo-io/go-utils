package debugutils

import (
	"io"
	"path/filepath"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/tarutils"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	aggregatorName = "aggregator"
	logsStr        = "logs"
	resourcesStr   = "resources"
)

type Aggregator struct {
	resourceCollector ResourceCollector
	logCollector      LogCollector
	storageClient     StorageClient
	fs                afero.Fs

	// root directory of the aggregated files
	dir string
}

func NewAggregator(resourceCollector ResourceCollector, logCollector LogCollector, storageClient StorageClient, fs afero.Fs, dir string) *Aggregator {
	return &Aggregator{resourceCollector: resourceCollector, logCollector: logCollector, storageClient: storageClient, fs: fs, dir: dir}
}

func DefaultAggregator() (*Aggregator, error) {
	fs := afero.NewOsFs()
	storageClient := NewFileStorageClient(fs)
	resourceCollector, err := DefaultResourceCollector()
	if err != nil {
		return nil, eris.Wrapf(err, "unable to initialize %s", aggregatorName)
	}
	logCollector, err := DefaultLogCollector()
	if err != nil {
		return nil, eris.Wrapf(err, "unable to initialize %s", aggregatorName)
	}
	tmpd, err := afero.TempDir(fs, "", "")
	if err != nil {
		return nil, err
	}
	return &Aggregator{
		logCollector:      logCollector,
		resourceCollector: resourceCollector,
		fs:                fs,
		dir:               tmpd,
		storageClient:     storageClient,
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
	if err := a.resourceCollector.SaveResources(a.storageClient, filepath.Join(a.dir, resourcesStr), kubeResources); err != nil {
		return err
	}
	logRequests, err := a.logCollector.GetLogRequests(unstructuredResources)
	if err != nil {
		return err
	}
	if err = a.logCollector.SaveLogs(a.storageClient, filepath.Join(a.dir, logsStr), logRequests); err != nil {
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
	if err := a.storageClient.Save(filepath.Dir(filename), &StorageObject{
		Name:     filepath.Base(filename),
		Resource: tarball,
	}); err != nil {
		return err
	}
	return nil
}

func (a *Aggregator) createSubResourceDirectories() error {
	directories := []string{resourcesStr, logsStr}
	for _, v := range directories {
		resourceDir := filepath.Join(a.dir, v)
		err := a.fs.Mkdir(resourceDir, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}
