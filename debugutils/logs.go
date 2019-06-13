package debugutils

import (
	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/go-utils/kubeutils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

type LogAggregator interface {
	LogsFromManifest(manifests helmchart.Manifests, opts metav1.ListOptions) ([]kuberesource.VersionedResources, error)
	RetrieveLogs(pods corev1.PodList, options corev1.PodLogOptions) ([]kuberesource.VersionedResources, error)
}

type ApiLogAggregator struct {
	clientset corev1client.CoreV1Interface
}

func NewApiLogAggregator() (*ApiLogAggregator, error) {
	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, err
	}
	clientset, err := corev1client.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &ApiLogAggregator{
		clientset: clientset,
	}, nil
}

// func (la *ApiLogAggregator) LogsFromManifest(manifests helmchart.Manifests, opts metav1.ListOptions) ([]kuberesource.VersionedResources, error) {
// 	resources, err := manifests.ResourceList()
// 	if err != nil {
// 		return nil, err
// 	}
//
// }
//
// func (la *ApiLogAggregator) RetrieveLogs(pods corev1.PodList, options corev1.PodLogOptions) ([]kuberesource.VersionedResources, error) {
// 	panic("implement me")
// }
