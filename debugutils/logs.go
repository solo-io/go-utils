package debugutils

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/installutils/kuberesource"
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

type LogsResponse struct {
	podMeta       metav1.ObjectMeta
	containerName string
	reader        io.ReadCloser
}

func (lr *LogsRequest) BuildId(dir string) string {
	return filepath.Join(dir, fmt.Sprintf("%s_%s_%s.log", lr.podMeta.Namespace, lr.podMeta.Name, lr.containerName))
}

func NewLogsRequest(podMeta metav1.ObjectMeta, containerName string, request *rest.Request) *LogsRequest {
	return &LogsRequest{podMeta: podMeta, containerName: containerName, request: request}
}

type LogStorageClient struct {
	fs  afero.Fs
	dir string
}

func NewLogFileStorage(fs afero.Fs, dir string) *LogStorageClient {
	return &LogStorageClient{fs: fs, dir: dir}
}

func (lfs *LogStorageClient) FetchLogs(requests []*LogsRequest) error {
	eg := errgroup.Group{}
	logsDir := filepath.Join(lfs.dir, "logs")
	err := lfs.fs.Mkdir(logsDir, 0777)
	if err != nil {
		return err
	}
	lfs.dir = logsDir
	for _, request := range requests {
		request := request
		eg.Go(func() error {
			reader, err := request.request.Stream()
			if err != nil {
				return err
			}
			defer reader.Close()
			file, err := lfs.fs.Create(request.BuildId(lfs.dir))
			if err != nil {
				return err
			}
			_, err = io.Copy(file, reader)
			return err
		})
	}
	return eg.Wait()
}

func (lfs *LogStorageClient) Dir() string {
	return lfs.dir
}

func (lfs *LogStorageClient) Clean() error {
	return lfs.fs.Remove(lfs.dir)
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
	resources, err := manifests.ResourceList()
	if err != nil {
		return nil, err
	}
	return lrb.LogsFromUnstructured(resources)
}

func (lrb *LogRequestBuilder) LogsFromUnstructured(resources kuberesource.UnstructuredResources) ([]*LogsRequest, error) {
	var result []*LogsRequest
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
	for _, v := range pod.Spec.InitContainers {
		opts := &corev1.PodLogOptions{
			Container: v.Name,
		}
		request := lrb.clientset.Pods(pod.Namespace).GetLogs(pod.Name, opts)
		result = append(result, NewLogsRequest(pod.ObjectMeta, v.Name, request))
	}
	return result
}
