package debugutils

import (
	"sync"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/stringutils"
	"golang.org/x/sync/errgroup"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

//go:generate mockgen -destination mocks_test.go -self_package github.com/solo-io/go-utils/debugutils -package debugutils github.com/solo-io/go-utils/debugutils PodFinder,LogCollector,ResourceCollector,StorageClient
//go:generate mockgen -destination mocks_kube_test.go  -package debugutils k8s.io/client-go/rest ResponseWrapper

type PodFinder interface {
	GetPods(resources kuberesource.UnstructuredResources) ([]*corev1.PodList, error)
}

const (
	labelPodFinderStr = "labelPodFinder"
)

type LabelPodFinder struct {
	client kubernetes.Interface
}

func NewLabelPodFinder(client kubernetes.Interface) *LabelPodFinder {
	return &LabelPodFinder{client: client}
}

func DefaultLabelPodFinder() (*LabelPodFinder, error) {
	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, errors.InitializationError(err, labelPodFinderStr)
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.InitializationError(err, labelPodFinderStr)
	}
	return &LabelPodFinder{
		client: client,
	}, nil
}

func (lpf *LabelPodFinder) GetPods(resources kuberesource.UnstructuredResources) ([]*corev1.PodList, error) {
	eg := errgroup.Group{}
	lock := sync.Mutex{}
	var result []*corev1.PodList
	for _, resource := range resources {
		resource := resource
		eg.Go(func() error {
			var list *corev1.PodList
			switch {
			case resource.GetKind() == "Pod":
				pod, err := lpf.client.CoreV1().Pods(resource.GetNamespace()).Get(resource.GetName(), metav1.GetOptions{})
				if err != nil {
					return err
				}
				list = &corev1.PodList{
					TypeMeta: metav1.TypeMeta{
						Kind:       "List",
						APIVersion: "v1",
					},
					Items: []corev1.Pod{*pod},
				}
			case stringutils.ContainsString(resource.GetKind(), ownerResources):
				matchLabels, err := handleOwnerResource(resource)
				if err != nil {
					return err
				}
				list, err = lpf.getPodsForMatchLabels(matchLabels, resource.GetNamespace())
				if err != nil {
					return err
				}
			default:
				return nil
			}
			lock.Lock()
			defer lock.Unlock()
			result = append(result, list)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

func (lpf *LabelPodFinder) getPodsForMatchLabels(matchLabels map[string]string, namespace string) (*corev1.PodList, error) {
	var set labels.Set = matchLabels
	return lpf.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: set.String(),
	})
}

func ConvertPodListsToUnstructured(pods []*corev1.PodList) (kuberesource.UnstructuredResources, error) {
	var result kuberesource.UnstructuredResources
	for _, list := range pods {
		convertedList, err := ConvertPodsToUnstructured(list)
		if err != nil {
			return nil, err
		}
		result = append(result, convertedList...)
	}
	return result, nil
}

func ConvertPodsToUnstructured(pods *corev1.PodList) (kuberesource.UnstructuredResources, error) {
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
	case *batchv1.Job:
		matchLabels = deploymentType.Spec.Selector.MatchLabels
	case *batchv1beta1.CronJob:
		matchLabels = deploymentType.Spec.JobTemplate.Spec.Selector.MatchLabels

	default:
		return nil, errors.Errorf("unable to determine the type of resource %v", obj)
	}
	return matchLabels, nil
}
