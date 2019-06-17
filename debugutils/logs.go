package debugutils

import (
	"fmt"
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

//go:generate mockgen -destination=./mocks/logs.go -source logs.go -package mocks

type LogCollector interface {
	GetLogRequests(resources kuberesource.UnstructuredResources) ([]*LogsRequest, error)
	SaveLogs(location string, requests []*LogsRequest) error
}

type LogsRequest struct {
	podMeta       metav1.ObjectMeta
	containerName string
	request       *rest.Request
}

func (lr *LogsRequest) ResourcePath(dir string) string {
	return filepath.Join(dir, lr.ResourceId())
}

func (lr *LogsRequest) ResourceId() string {
	return fmt.Sprintf("%s_%s_%s.log", lr.podMeta.Namespace, lr.podMeta.Name, lr.containerName)
}

func NewLogsRequest(podMeta metav1.ObjectMeta, containerName string, request *rest.Request) *LogsRequest {
	return &LogsRequest{podMeta: podMeta, containerName: containerName, request: request}
}

type logCollector struct {
	logRequestBuilder *LogRequestBuilder
	storageClient     StorageClient
}

func NewLogCollector(logRequestBuilder *LogRequestBuilder, storageClient StorageClient) *logCollector {
	return &logCollector{logRequestBuilder: logRequestBuilder, storageClient: storageClient}

}

func DefaultLogCollector() (*logCollector, error) {
	logRequestBuilder, err := DefaultLogRequestBuilder()
	if err != nil {
		return nil, err
	}
	storageClient := NewFileStorageClient(afero.NewOsFs())
	return &logCollector{
		storageClient:     storageClient,
		logRequestBuilder: logRequestBuilder,
	}, nil
}

func (lc *logCollector) GetLogRequestsFromManifest(manifests helmchart.Manifests) ([]*LogsRequest, error) {
	resources, err := manifests.ResourceList()
	if err != nil {
		return nil, err
	}
	return lc.logRequestBuilder.LogsFromUnstructured(resources)
}

func (lc *logCollector) GetLogRequests(resources kuberesource.UnstructuredResources) ([]*LogsRequest, error) {
	return lc.logRequestBuilder.LogsFromUnstructured(resources)
}

func (lc *logCollector) SaveLogs(location string, requests []*LogsRequest) error {
	eg := errgroup.Group{}
	for _, request := range requests {
		// necessary to shadow this variable so that it is unique within the goroutine
		restRequest := request
		eg.Go(func() error {
			reader, err := restRequest.request.Stream()
			if err != nil {
				return err
			}
			defer reader.Close()
			return lc.storageClient.Save(location, &StorageObject{
				resource: reader,
				name: restRequest.ResourceId(),
			})
		})
	}
	return eg.Wait()
}

type LogRequestBuilder struct {
	clientset corev1client.CoreV1Interface
	podFinder PodFinder
}

func NewLogRequestBuilder(clientset corev1client.CoreV1Interface, podFinder PodFinder) *LogRequestBuilder {
	return &LogRequestBuilder{clientset: clientset, podFinder: podFinder}
}

func DefaultLogRequestBuilder() (*LogRequestBuilder, error) {
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
