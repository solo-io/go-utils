package debugutils

import (
	"fmt"
	"io"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type LogMeta struct {
	PodMeta       metav1.ObjectMeta
	ContainerName string
}

type LogsRequest struct {
	LogMeta
	Request rest.ResponseWrapper
}

type LogsResponse struct {
	LogMeta
	Response io.ReadCloser
}

func (lr LogMeta) ResourcePath(dir string) string {
	return filepath.Join(dir, lr.ResourceId())
}

func (lr LogMeta) ResourceId() string {
	return fmt.Sprintf("%s_%s_%s.log", lr.PodMeta.Namespace, lr.PodMeta.Name, lr.ContainerName)
}

func NewLogsRequest(podMeta metav1.ObjectMeta, containerName string, request *rest.Request) *LogsRequest {
	return &LogsRequest{LogMeta: LogMeta{PodMeta: podMeta, ContainerName: containerName}, Request: request}
}
