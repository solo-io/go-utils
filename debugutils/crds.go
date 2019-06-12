package debugutils

import (
	"encoding/json"
	"sync"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/stringutils"
	"golang.org/x/sync/errgroup"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var (
	initializationError = func(err error) error {
		return errors.Wrapf(err, "unable to initialize resourceCollector")
	}
)

type resourceCollector struct {
	client        kubernetes.Interface
	dynamicClient dynamic.Interface
	restMapper    meta.RESTMapper
}

func NewCrdCollector() (*resourceCollector, error) {
	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, initializationError(err)
	}
	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, initializationError(err)
	}
	restMapper, err := apiutil.NewDiscoveryRESTMapper(cfg)
	if err != nil {
		return nil, initializationError(err)
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, initializationError(err)
	}
	return &resourceCollector{
		dynamicClient: dynamicClient,
		restMapper:    restMapper,
		client:        client,
	}, nil
}



func (cc *resourceCollector) ResourcesFromManifest(manifests helmchart.Manifests, opts metav1.ListOptions) ([]kuberesource.VersionedResources, error) {
	resources, err := manifests.ResourceList()
	if err != nil {
		return nil, err
	}
	return cc.RetrieveResources(resources, "", opts)
}

func (cc *resourceCollector) RetrieveResources(resources kuberesource.UnstructuredResources, namespace string, opts metav1.ListOptions) ([]kuberesource.VersionedResources, error) {
	var result kuberesource.UnstructuredResources
	eg := errgroup.Group{}
	lock := sync.RWMutex{}

	for _, kind := range resources {
		kind := kind
		eg.Go(func() error {
			resources, err := cc.handleUnstructuredResource(kind, namespace, opts)
			if err != nil {
				return err
			}
			lock.Lock()
			defer lock.Unlock()
			result = append(result, resources...)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return result.GroupedByGVK(), nil
}

var ownerResources = []string{"Deployment", "DaemonSet", "Job", "CronJob"}

func (cc *resourceCollector) handleUnstructuredResource(resource *unstructured.Unstructured, namespace string, opts metav1.ListOptions) (kuberesource.UnstructuredResources, error) {
	switch  {
	case resource.GetKind() == "CustomResourceDefinition":
		return cc.listAllFromNamespace(resource, namespace, opts)
	case stringutils.ContainsString(resource.GetKind(), ownerResources):
		matchLabels, err := handleOwnerResource(resource)
		if err != nil {
			return nil, err
		}
		return cc.getPodsForMatchLabels(matchLabels, namespace)
	default:
		return cc.getResource(resource)
	}
}

func (cc *resourceCollector) getPodsForMatchLabels(matchLabels map[string]string, namespace string) (kuberesource.UnstructuredResources, error) {
	var set labels.Set = matchLabels
	pods, err := cc.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: set.String(),
	})
	if err != nil {
		return nil, err
	}
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

func (cc *resourceCollector) listAllFromNamespace(resource *unstructured.Unstructured, namespace string, opts metav1.ListOptions) (kuberesource.UnstructuredResources, error) {
	kind, err := cc.gvrFromUnstructured(*resource)
	if err != nil {
		return nil, err
	}
	list, err := cc.dynamicClient.Resource(kind).Namespace(namespace).List(opts)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to retrieve resources for kind %v", kind)
	}
	result := make(kuberesource.UnstructuredResources, len(list.Items))
	for idx, val := range list.Items {
		result[idx] = &val
	}
	return result, nil
}

func (cc *resourceCollector) getResource(resource *unstructured.Unstructured) (kuberesource.UnstructuredResources, error) {
	kind, err := cc.gvrFromUnstructured(*resource)
	if err != nil {
		return nil, err
	}
	var res *unstructured.Unstructured
	if resource.GetNamespace() != "" {
		res, err = cc.dynamicClient.Resource(kind).Namespace(resource.GetNamespace()).Get(resource.GetName(), metav1.GetOptions{})
	} else {
		res, err = cc.dynamicClient.Resource(kind).Get(resource.GetName(), metav1.GetOptions{})

	}
	if err != nil {
		return nil, errors.Wrapf(err, "unable to retrieve resources for kind %v", kind)
	}
	return kuberesource.UnstructuredResources{res}, nil
}

func (cc *resourceCollector) gvrFromUnstructured(resource unstructured.Unstructured) (schema.GroupVersionResource, error) {
	gvr := schema.GroupVersionResource{
		Group:    resource.GetObjectKind().GroupVersionKind().Group,
		Version:  resource.GetObjectKind().GroupVersionKind().Version,
		Resource: resource.GetKind(),
	}
	if gvr.Resource == "CustomResourceDefinition" {
		var cdr apiextensions.CustomResourceDefinition
		rawJson, err := json.Marshal(resource.Object)
		if err != nil {
			return schema.GroupVersionResource{}, err
		}
		if err := json.Unmarshal(rawJson, &cdr); err != nil {
			return schema.GroupVersionResource{}, err
		}
		gvr = schema.GroupVersionResource{
			Group:    cdr.Spec.Group,
			Version:  cdr.Spec.Version,
			Resource: cdr.Spec.Names.Plural,
		}
	}
	result, err := cc.restMapper.ResourceFor(gvr)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	return result, nil
}


func filterCrds(resources kuberesource.UnstructuredResources) ([]*apiextensions.CustomResourceDefinition, error) {
	var result []*apiextensions.CustomResourceDefinition
	for _, resource := range resources {
		runtimeObj, err := kuberesource.ConvertUnstructured(resource)
		if err != nil {
			return nil, err
		}
		crd, ok := runtimeObj.(*apiextensions.CustomResourceDefinition)
		if ok {
			result = append(result, crd)
		}
	}

	return result, nil
}
