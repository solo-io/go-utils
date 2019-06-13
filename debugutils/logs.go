package debugutils

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/spf13/afero"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type LogsRequest struct {
	podMeta       metav1.ObjectMeta
	containerName string
	request       *rest.Request
}

func (lr *LogsRequest) BuildFileName(dir string) string {
	return filepath.Join(dir, fmt.Sprintf("%s.%s.%s", lr.podMeta.Namespace, lr.podMeta.Name, lr.containerName))
}

func NewLogsRequest(podMeta metav1.ObjectMeta, containerName string, request *rest.Request) *LogsRequest {
	return &LogsRequest{podMeta: podMeta, containerName: containerName, request: request}
}

type LogStorage interface {
	SaveLogs(requests []*LogsRequest) error
}

type logFileStorage struct {
	fs  afero.Fs
	dir string
}

func NewLogFileStorage(fs afero.Fs, dir string) *logFileStorage {
	return &logFileStorage{fs: fs, dir: dir}
}

func (lfs *logFileStorage) SaveLogs(requests []*LogsRequest) error {
	eg := errgroup.Group{}
	for _, request := range requests {
		request := request
		eg.Go(func() error {
			respbyt, err := request.request.DoRaw()
			if err != nil {
				return err
			}
			file, err := lfs.fs.Create(request.BuildFileName(lfs.dir))
			if err != nil {
				return err
			}
			return ioutil.WriteFile(file.Name(), respbyt, 0777)
		})
	}
	return eg.Wait()
}

type LogRequestBuilder struct {
	clientset corev1client.CoreV1Interface
	podFinder PodFinder
}

func NewLogRequestBuilder() (*LogRequestBuilder, error) {
	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, err
	}
	clientset, err := corev1client.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	podFinder, err := NewLabelPodFinder()
	if err != nil {
		return nil, err
	}
	return &LogRequestBuilder{
		clientset: clientset,
		podFinder: podFinder,
	}, nil
}

func (lrb *LogRequestBuilder) LogsFromManifest(manifests helmchart.Manifests) ([]*LogsRequest, error) {
	var result []*LogsRequest
	resources, err := manifests.ResourceList()
	if err != nil {
		return nil, err
	}
	pods, err := lrb.podFinder.GetPods(resources)
	if err != nil {
		return nil, err
	}
	for _, v := range pods {
		result = append(result, lrb.RetrieveLogs(v)...)
	}
	return result, nil
}

func (lrb *LogRequestBuilder) RetrieveLogs(pods *corev1.PodList) []*LogsRequest {
	var result []*LogsRequest
	for _, v := range pods.Items {
		result = append(result, lrb.buildLogsRequest(v)...)
	}
	return result
}

func (lrb *LogRequestBuilder) buildLogsRequest(pod corev1.Pod) []*LogsRequest {
	var result []*LogsRequest
	for _, v := range pod.Spec.Containers {
		opts := &corev1.PodLogOptions{
			Container: v.Name,
		}
		request := lrb.clientset.Pods(pod.Namespace).GetLogs(pod.Name, opts)
		result = append(result, NewLogsRequest(pod.ObjectMeta, v.Name, request))
	}
	for _, v := range pod.Spec.Containers {
		opts := &corev1.PodLogOptions{
			Container: v.Name,
		}
		request := lrb.clientset.Pods(pod.Namespace).GetLogs(pod.Name, opts)
		result = append(result, NewLogsRequest(pod.ObjectMeta, v.Name, request))
	}
	return result
}
