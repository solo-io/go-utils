package common

import (
	"fmt"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type LogsRequest struct {
	PodMeta       metav1.ObjectMeta
	ContainerName string
	Request       rest.ResponseWrapper
}

func (lr *LogsRequest) ResourcePath(dir string) string {
	return filepath.Join(dir, lr.ResourceId())
}

func (lr *LogsRequest) ResourceId() string {
	return fmt.Sprintf("%s_%s_%s.log", lr.PodMeta.Namespace, lr.PodMeta.Name, lr.ContainerName)
}

func NewLogsRequest(podMeta metav1.ObjectMeta, containerName string, request *rest.Request) *LogsRequest {
	return &LogsRequest{PodMeta: podMeta, ContainerName: containerName, Request: request}
}
