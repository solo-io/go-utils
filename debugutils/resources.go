package debugutils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/spf13/afero"
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

var (
	initializationError = func(err error, obj string) error {
		return errors.Wrapf(err, "unable to initialize %s", obj)
	}
)

type ResourceCollector interface {
	ResourcesFromManifest(manifests helmchart.Manifests, opts metav1.ListOptions) ([]kuberesource.VersionedResources, error)
	RetrieveResources(resources kuberesource.UnstructuredResources, namespace string, opts metav1.ListOptions) ([]kuberesource.VersionedResources, error)
	SaveResources(versionedResources []kuberesource.VersionedResources, fs afero.Fs, dir string) error
}

type resourceCollector struct {
	dynamicClient dynamic.Interface
	restMapper    meta.RESTMapper
	podFinder     PodFinder
}

func NewResourceCollector() (*resourceCollector, error) {
	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, initializationError(err, resourceCollectorStr)
	}
	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, initializationError(err, resourceCollectorStr)
	}
	restMapper, err := apiutil.NewDiscoveryRESTMapper(cfg)
	if err != nil {
		return nil, initializationError(err, resourceCollectorStr)
	}
	podFinder, err := NewLabelPodFinder()
	if err != nil {
		return nil, initializationError(err, resourceCollectorStr)
	}
	return &resourceCollector{
		dynamicClient: dynamicClient,
		restMapper:    restMapper,
		podFinder:     podFinder,
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
	pods, err := cc.podFinder.GetPods(resources)
	if err != nil {
		return nil, err
	}
	convertedPods, err := convertPodListsToUnstructured(pods)
	if err != nil {
		return nil, err
	}
	result = append(result, convertedPods...)
	return result.GroupedByGVK(), nil
}

var ownerResources = []string{"Deployment", "DaemonSet", "Job", "CronJob"}

func (cc *resourceCollector) handleUnstructuredResource(resource *unstructured.Unstructured, namespace string, opts metav1.ListOptions) (kuberesource.UnstructuredResources, error) {
	switch {
	case resource.GetKind() == "CustomResourceDefinition":
		return cc.listAllFromNamespace(resource, namespace, opts)
	default:
		return cc.getResource(resource)
	}
}

func (cc *resourceCollector) listAllFromNamespace(resource *unstructured.Unstructured, namespace string, opts metav1.ListOptions) (kuberesource.UnstructuredResources, error) {
	kind, err := cc.gvrFromUnstructured(*resource)
	if err != nil {
		return nil, err
	}
	var list *unstructured.UnstructuredList
	if namespace == "" {
		list, err = cc.dynamicClient.Resource(kind).List(opts)
	} else {
		list, err = cc.dynamicClient.Resource(kind).Namespace(namespace).List(opts)
	}

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

func (cc *resourceCollector) SaveResources(versionedResources []kuberesource.VersionedResources, fs afero.Fs, dir string) error {
	resourceDir := filepath.Join(dir, "resources")
	err := fs.Mkdir(resourceDir, 0777)
	if err != nil {
		return err
	}
	for _, versionedResource := range versionedResources {
		fileName := filepath.Join(resourceDir, fmt.Sprintf("%s_%s.yaml", versionedResource.GVK.Kind, versionedResource.GVK.Version))
		_, err := fs.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
		if err != nil {
			return err
		}
		tmpManifests, err := helmchart.ManifestsFromResources(versionedResource.Resources)
		if err != nil {
			return err
		}
		err = afero.WriteFile(fs, fileName, []byte(tmpManifests.CombinedString()), 0777)
		if err != nil {
			return err
		}
	}
	return nil
}
