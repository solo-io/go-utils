package debugutils

import (
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/kubernetes/pkg/apis/batch"
)

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
