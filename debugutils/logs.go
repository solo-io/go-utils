package debugutils

import (
	"github.com/solo-io/go-utils/debugutils/common"
	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/go-utils/kubeutils"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

type LogCollector interface {
	GetLogRequests(resources kuberesource.UnstructuredResources) ([]*common.LogsRequest, error)
	SaveLogs(client common.StorageClient, location string, requests []*common.LogsRequest) error
}


type logCollector struct {
	logRequestBuilder *LogRequestBuilder
}

func NewLogCollector(logRequestBuilder *LogRequestBuilder) *logCollector {
	return &logCollector{logRequestBuilder: logRequestBuilder}

}

func DefaultLogCollector() (*logCollector, error) {
	logRequestBuilder, err := DefaultLogRequestBuilder()
	if err != nil {
		return nil, err
	}
	return &logCollector{
		logRequestBuilder: logRequestBuilder,
	}, nil
}

func (lc *logCollector) GetLogRequestsFromManifest(manifests helmchart.Manifests) ([]*common.LogsRequest, error) {
	resources, err := manifests.ResourceList()
	if err != nil {
		return nil, err
	}
	return lc.logRequestBuilder.LogsFromUnstructured(resources)
}

func (lc *logCollector) GetLogRequests(resources kuberesource.UnstructuredResources) ([]*common.LogsRequest, error) {
	return lc.logRequestBuilder.LogsFromUnstructured(resources)
}

func (lc *logCollector) SaveLogs(storageClient common.StorageClient, location string, requests []*common.LogsRequest) error {
	eg := errgroup.Group{}
	for _, request := range requests {
		// necessary to shadow this variable so that it is unique within the goroutine
		restRequest := request
		eg.Go(func() error {
			reader, err := restRequest.Request.Stream()
			if err != nil {
				return err
			}
			defer reader.Close()
			return storageClient.Save(location, &common.StorageObject{
				Resource: reader,
				Name: restRequest.ResourceId(),
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
	podFinder, err := DefaultLabelPodFinder()
	if err != nil {
		return nil, err
	}
	return &LogRequestBuilder{
		clientset: clientset,
		podFinder: podFinder,
	}, nil
}

func (lrb *LogRequestBuilder) LogsFromUnstructured(resources kuberesource.UnstructuredResources) ([]*common.LogsRequest, error) {
	var result []*common.LogsRequest
	pods, err := lrb.podFinder.GetPods(resources)
	if err != nil {
		return nil, err
	}
	for _, v := range pods {
		result = append(result, lrb.RetrieveLogs(v)...)
	}
	return result, nil
}

func (lrb *LogRequestBuilder) RetrieveLogs(pods *corev1.PodList) []*common.LogsRequest {
	var result []*common.LogsRequest
	for _, v := range pods.Items {
		result = append(result, lrb.buildLogsRequest(v)...)
	}
	return result
}

func (lrb *LogRequestBuilder) buildLogsRequest(pod corev1.Pod) []*common.LogsRequest {
	var result []*common.LogsRequest
	for _, v := range pod.Spec.Containers {
		opts := &corev1.PodLogOptions{
			Container: v.Name,
		}
		request := lrb.clientset.Pods(pod.Namespace).GetLogs(pod.Name, opts)
		result = append(result, common.NewLogsRequest(pod.ObjectMeta, v.Name, request))
	}
	for _, v := range pod.Spec.InitContainers {
		opts := &corev1.PodLogOptions{
			Container: v.Name,
		}
		request := lrb.clientset.Pods(pod.Namespace).GetLogs(pod.Name, opts)
		result = append(result, common.NewLogsRequest(pod.ObjectMeta, v.Name, request))
	}
	return result
}
