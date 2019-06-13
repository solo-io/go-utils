package debugutils

import (
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/stringutils"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/apis/batch"
)

type PodFinder interface {
	GetPods(resources kuberesource.UnstructuredResources) ([]*corev1.PodList, error)
}

const (
	labelPodFinderStr = "labelPodFinder"
)

type LabelPodFinder struct {
	client kubernetes.Interface
}

func NewLabelPodFinder() (*LabelPodFinder, error) {
	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, initializationError(err, labelPodFinderStr)
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, initializationError(err, labelPodFinderStr)
	}
	return &LabelPodFinder{
		client: client,
	}, nil
}

func (mpf *LabelPodFinder) GetPods(resources kuberesource.UnstructuredResources) ([]*corev1.PodList, error) {
	var result []*corev1.PodList
	for _, resource := range  resources {
		var matchLabels map[string]string
		var err error
		switch {
		case resource.GetKind() == "Pod":
			matchLabels = resource.GetLabels()
		case stringutils.ContainsString(resource.GetKind(), ownerResources):
			matchLabels, err = handleOwnerResource(resource)
			if err != nil {
				return nil, err
			}
		default:
			continue
		}
		res, err := mpf.getPodsForMatchLabels(matchLabels, resource.GetNamespace())
		if err != nil {
			return nil, err
		}
		result = append(result, res)
	}

	return result, nil
}

func (mpf *LabelPodFinder) getPodsForMatchLabels(matchLabels map[string]string, namespace string) (*corev1.PodList, error) {
	var set labels.Set = matchLabels
	return mpf.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: set.String(),
	})
}

func convertPodListsToUnstructured(pods []*corev1.PodList) (kuberesource.UnstructuredResources, error) {
	var result kuberesource.UnstructuredResources
	for _, list := range pods {
		convertedList, err := convertPodsToUnstructured(list)
		if err != nil {
			return nil, err
		}
		result = append(result, convertedList...)
	}
	return result, nil
}

func convertPodsToUnstructured(pods *corev1.PodList) (kuberesource.UnstructuredResources, error) {
	result := make(kuberesource.UnstructuredResources, len(pods.Items))
	for idx, val := range pods.Items {
		resource, err := kuberesource.ConvertToUnstructured(&val)
		if err != nil {
			return nil, err
		}
		resource.SetKind("Pod")
		resource.SetAPIVersion("v1")
		result[idx] = resource
	}
	return result, nil
}

func handleOwnerResource(resource *unstructured.Unstructured) (map[string]string, error) {
	obj, err := kuberesource.ConvertUnstructured(resource)
	if err != nil {
		return nil, err
	}
	var matchLabels map[string]string
	switch deploymentType := obj.(type) {
	case *extensionsv1beta1.Deployment:
		matchLabels = deploymentType.Spec.Selector.MatchLabels
	case *appsv1.Deployment:
		matchLabels = deploymentType.Spec.Selector.MatchLabels
	case *appsv1beta2.Deployment:
		matchLabels = deploymentType.Spec.Selector.MatchLabels
	case *extensionsv1beta1.DaemonSet:
		matchLabels = deploymentType.Spec.Selector.MatchLabels
	case *appsv1.DaemonSet:
		matchLabels = deploymentType.Spec.Selector.MatchLabels
	case *appsv1beta2.DaemonSet:
		matchLabels = deploymentType.Spec.Selector.MatchLabels
	case *batch.Job:
		matchLabels = deploymentType.Spec.Selector.MatchLabels
	case *batch.CronJob:
		matchLabels = deploymentType.Spec.JobTemplate.Spec.Selector.MatchLabels

	default:
		return nil, errors.Errorf("unable to determine the type of resource %v", obj)
	}
	return matchLabels, nil
}
