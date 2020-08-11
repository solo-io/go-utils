package debugutils

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/go-utils/kubeutils"
	"golang.org/x/sync/errgroup"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

const (
	resourceCollectorStr = "resourceCollector"
)

type ResourceCollector interface {
	RetrieveResources(ctx context.Context, resources kuberesource.UnstructuredResources, namespace string, opts metav1.ListOptions) ([]kuberesource.VersionedResources, error)
	SaveResources(ctx context.Context, client StorageClient, location string, versionedResources []kuberesource.VersionedResources) error
}

type resourceCollector struct {
	dynamicClient dynamic.Interface
	restMapper    meta.RESTMapper
	podFinder     PodFinder
}

func DefaultResourceCollector() (*resourceCollector, error) {
	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, eris.Wrapf(err, "unable to initialize %s", resourceCollectorStr)
	}
	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, eris.Wrapf(err, "unable to initialize %s", resourceCollectorStr)
	}
	restMapper, err := apiutil.NewDiscoveryRESTMapper(cfg)
	if err != nil {
		return nil, eris.Wrapf(err, "unable to initialize %s", resourceCollectorStr)
	}
	podFinder, err := DefaultLabelPodFinder()
	if err != nil {
		return nil, eris.Wrapf(err, "unable to initialize %s", resourceCollectorStr)
	}
	return &resourceCollector{
		dynamicClient: dynamicClient,
		restMapper:    restMapper,
		podFinder:     podFinder,
	}, nil
}

func (rc *resourceCollector) RetrieveResourcesFromManifest(ctx context.Context, manifests helmchart.Manifests, opts metav1.ListOptions) ([]kuberesource.VersionedResources, error) {
	resources, err := manifests.ResourceList()
	if err != nil {
		return nil, err
	}
	return rc.RetrieveResources(ctx, resources, "", opts)
}

func (rc *resourceCollector) RetrieveResources(ctx context.Context, resources kuberesource.UnstructuredResources, namespace string, opts metav1.ListOptions) ([]kuberesource.VersionedResources, error) {
	var result kuberesource.UnstructuredResources
	eg := errgroup.Group{}
	lock := sync.RWMutex{}

	for _, kind := range resources {
		kind := kind
		eg.Go(func() error {
			resources, err := rc.handleUnstructuredResource(ctx, kind, namespace, opts)
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
	pods, err := rc.podFinder.GetPods(ctx, resources)
	if err != nil {
		return nil, err
	}
	convertedPods, err := ConvertPodListsToUnstructured(pods)
	if err != nil {
		return nil, err
	}
	result = append(result, convertedPods...)
	return result.GroupedByGVK(), nil
}

var ownerResources = []string{"Deployment", "DaemonSet", "Job", "CronJob"}

func (rc *resourceCollector) handleUnstructuredResource(ctx context.Context, resource *unstructured.Unstructured, namespace string, opts metav1.ListOptions) (kuberesource.UnstructuredResources, error) {
	switch {
	case resource.GetKind() == "CustomResourceDefinition":
		return rc.listAllFromNamespace(ctx, resource, namespace, opts)
	default:
		return rc.getResource(ctx, resource)
	}
}

func (rc *resourceCollector) listAllFromNamespace(ctx context.Context, resource *unstructured.Unstructured, namespace string, opts metav1.ListOptions) (kuberesource.UnstructuredResources, error) {
	kind, err := rc.gvrFromUnstructured(*resource)
	if err != nil {
		return nil, err
	}
	var list *unstructured.UnstructuredList
	if namespace == "" {
		list, err = rc.dynamicClient.Resource(kind).List(ctx, opts)
	} else {
		list, err = rc.dynamicClient.Resource(kind).Namespace(namespace).List(ctx, opts)
	}

	if err != nil {
		return nil, eris.Wrapf(err, "unable to retrieve resources for kind %v", kind)
	}
	result := make(kuberesource.UnstructuredResources, len(list.Items))
	for idx, val := range list.Items {
		result[idx] = &val
	}
	return result, nil
}

func (rc *resourceCollector) getResource(ctx context.Context, resource *unstructured.Unstructured) (kuberesource.UnstructuredResources, error) {
	kind, err := rc.gvrFromUnstructured(*resource)
	if err != nil {
		return nil, err
	}
	var res *unstructured.Unstructured
	if resource.GetNamespace() != "" {
		res, err = rc.dynamicClient.Resource(kind).Namespace(resource.GetNamespace()).Get(ctx, resource.GetName(), metav1.GetOptions{})
	} else {
		res, err = rc.dynamicClient.Resource(kind).Get(ctx, resource.GetName(), metav1.GetOptions{})

	}
	if err != nil {
		return nil, eris.Wrapf(err, "unable to retrieve resources for kind %v", kind)
	}
	return kuberesource.UnstructuredResources{res}, nil
}

func (rc *resourceCollector) gvrFromUnstructured(resource unstructured.Unstructured) (schema.GroupVersionResource, error) {
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
	result, err := rc.restMapper.ResourceFor(gvr)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	return result, nil
}

func (rc *resourceCollector) SaveResources(storageClient StorageClient, location string, versionedResources []kuberesource.VersionedResources) error {
	var storageObjects []*StorageObject
	for _, versionedResource := range versionedResources {
		tmpManifests, err := helmchart.ManifestsFromResources(versionedResource.Resources)
		if err != nil {
			return err
		}
		reader := strings.NewReader(tmpManifests.CombinedString())
		resourceName := fmt.Sprintf("%s_%s.yaml", versionedResource.GVK.Kind, versionedResource.GVK.Version)
		storageObjects = append(storageObjects, &StorageObject{
			Resource: reader,
			Name:     resourceName,
		})
	}
	return storageClient.Save(location, storageObjects...)
}
