package debugutils

import (
	"sync"
	"time"

	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/go-utils/kubeutils"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

type LogCollector interface {
	GetLogRequests(resources kuberesource.UnstructuredResources) ([]*LogsRequest, error)
	SaveLogs(client StorageClient, location string, requests []*LogsRequest) error
	StreamLogs(requests []*LogsRequest) ([]*LogsResponse, error)
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

func (lc *logCollector) SaveLogs(storageClient StorageClient, location string, requests []*LogsRequest) error {
	eg := errgroup.Group{}
	responses, err := lc.StreamLogs(requests)
	if err != nil {
		return err
	}
	for _, response := range responses {
		response := response
		eg.Go(func() error {
			defer response.Response.Close()
			return storageClient.Save(location, &StorageObject{
				Resource: response.Response,
				Name:     response.ResourceId(),
			})
		})
	}
	return eg.Wait()
}

func (lc *logCollector) StreamLogs(requests []*LogsRequest) ([]*LogsResponse, error) {
	result := make([]*LogsResponse, 0, len(requests))
	eg := errgroup.Group{}
	lock := sync.Mutex{}
	for _, request := range requests {
		// necessary to shadow this variable so that it is unique within the goroutine
		restRequest := request
		eg.Go(func() error {
			reader, err := restRequest.Request.Stream()
			if err != nil {
				return err
			}
			lock.Lock()
			defer lock.Unlock()
			result = append(result, &LogsResponse{
				LogMeta: LogMeta{
					PodMeta:       restRequest.PodMeta,
					ContainerName: restRequest.ContainerName,
				},
				Response: reader,
			})
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

type LogRequestBuilder struct {
	clientset corev1client.CoreV1Interface
	podFinder PodFinder
}

type LogRequestOptions func(options *corev1.PodLogOptions)

var (
	FollowLogs LogRequestOptions = func(options *corev1.PodLogOptions) {
		options.Follow = true
	}
	PreviousLogs LogRequestOptions = func(options *corev1.PodLogOptions) {
		options.Previous = true

	}
	LogsSince = func(since time.Time) LogRequestOptions {
		return func(options *corev1.PodLogOptions) {
			options.SinceTime = &metav1.Time{Time: since,}
		}
	}
)

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

func (lrb *LogRequestBuilder) LogsFromUnstructured(resources kuberesource.UnstructuredResources, opts ...LogRequestOptions) ([]*LogsRequest, error) {
	var result []*LogsRequest
	pods, err := lrb.podFinder.GetPods(resources)
	if err != nil {
		return nil, err
	}
	for _, v := range pods {
		result = append(result, lrb.RetrieveLogs(v, opts...)...)
	}
	return result, nil
}

func (lrb *LogRequestBuilder) RetrieveLogs(pods *corev1.PodList, opts ...LogRequestOptions) []*LogsRequest {
	var result []*LogsRequest
	for _, v := range pods.Items {
		result = append(result, lrb.buildLogsRequest(v)...)
	}
	return result
}

func (lrb *LogRequestBuilder) buildLogsRequest(pod corev1.Pod, optsFunc ...LogRequestOptions) []*LogsRequest {
	var result []*LogsRequest
	opts := &corev1.PodLogOptions{}
	for _, f := range optsFunc {
		if f != nil {
			f(opts)
		}
	}
	for _, v := range pod.Spec.Containers {
		opts.Container = v.Name
		request := lrb.clientset.Pods(pod.Namespace).GetLogs(pod.Name, opts)
		result = append(result, NewLogsRequest(pod.ObjectMeta, v.Name, request))
	}
	for _, v := range pod.Spec.InitContainers {
		opts.Container = v.Name
		request := lrb.clientset.Pods(pod.Namespace).GetLogs(pod.Name, opts)
		result = append(result, NewLogsRequest(pod.ObjectMeta, v.Name, request))
	}
	return result
}
