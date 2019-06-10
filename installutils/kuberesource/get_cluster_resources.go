package kuberesource

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/goph/emperror"
	"github.com/solo-io/go-utils/contextutils"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type clusterResourceFetcher struct {
	cfg         *rest.Config
	filterFuncs []FilterResource
	namespaces  []string
}

func NewClusterResourceFetcher(cfg *rest.Config, namespaces []string, filterFuncs ...FilterResource) ClusterResourceFetcher {
	return &clusterResourceFetcher{
		cfg:         cfg,
		namespaces:  namespaces,
		filterFuncs: filterFuncs,
	}
}

type FilterResource func(schema.GroupVersionResource) bool

type ClusterResourceFetcher interface {
	GetClusterResources(ctx context.Context) (UnstructuredResources, error)
}

func (c *clusterResourceFetcher) GetClusterResources(ctx context.Context) (UnstructuredResources, error) {
	// discovery client
	disc, err := discovery.NewDiscoveryClientForConfig(c.cfg)
	if err != nil {
		return nil, err
	}

	// resource client
	client, err := dynamic.NewForConfig(c.cfg)
	if err != nil {
		return nil, err
	}

	// list api resources that can be CRUD'ed
	serverResources, err := disc.ServerResources()
	if err != nil {
		return nil, err
	}
	crudableResources := discovery.FilteredBy(discovery.SupportsAllVerbs{Verbs: []string{"create", "list", "watch", "delete"}}, serverResources)

	gv, err := discovery.GroupVersionResources(crudableResources)
	if err != nil {
		return nil, err
	}
	// convert map to slice
	var groupVersionResources []schema.GroupVersionResource
	for res := range gv {
		groupVersionResources = append(groupVersionResources, res)
	}

	var writeAccess sync.Mutex
	var allResources UnstructuredResources
	var g errgroup.Group
	for _, gvr := range filterGroupVersions(groupVersionResources, c.filterFuncs...) {
		gvr := gvr
		g.Go(func() error {
			contextutils.LoggerFrom(ctx).Infow("listing all", "resourceType", gvr)

			if c.namespaces == nil {
				resources, err := client.Resource(gvr).List(metav1.ListOptions{})
				if err != nil {
					return emperror.With(err, "group_version_resource", gvr)
				}

				for i := range resources.Items {
					res := &resources.Items[i]
					contextutils.LoggerFrom(ctx).Debugf("appending %s: %v.%v", res.GroupVersionKind(), res.GetNamespace(), res.GetName())
					writeAccess.Lock()
					allResources = append(allResources, res)
					writeAccess.Unlock()
				}
			} else {
				for _, ns := range c.namespaces {
					resources, err := client.Resource(gvr).Namespace(ns).List(metav1.ListOptions{})
					if err != nil {
						contextutils.LoggerFrom(ctx).Errorw("Error getting resources", zap.Error(err), zap.String("ns", ns))
						return emperror.With(err, "group_version_resource", gvr)
					}
					for i := range resources.Items {
						res := &resources.Items[i]
						contextutils.LoggerFrom(ctx).Debugf("appending %s: %v.%v", res.GroupVersionKind(), res.GetNamespace(), res.GetName())
						writeAccess.Lock()
						allResources = append(allResources, res)
						writeAccess.Unlock()
					}
				}
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error waiting for cache init", zap.Error(err))
		return nil, err
	}

	return allResources.Sort(), nil
}

/*
Use to get all CRUD'able resources for a cluster
Returns a list of UnstructuredResources which are a wrapper for the
map[string]interface{} type created from Kubernetes-style JSON objects.

Warning: slow function, be sure to call asynchronously.
Filter funcs can be passed to reduce the latency of this function,
i.e.: by excluding resource types (each resource type gets its own
query, contributing to latency of this function).
*/
func GetClusterResources(ctx context.Context, cfg *rest.Config, filterFuncs ...FilterResource) (UnstructuredResources, error) {
	return NewClusterResourceFetcher(cfg, nil, filterFuncs...).GetClusterResources(ctx)
}

func filterGroupVersions(groupVersions []schema.GroupVersionResource, filterFuncs ...FilterResource) []schema.GroupVersionResource {
	var filteredGroupVersions []schema.GroupVersionResource
	for _, resourceType := range groupVersions {
		var filtered bool
		for _, filter := range filterFuncs {
			if filter(resourceType) {
				filtered = true
				break
			}
		}
		if filtered {
			continue
		}
		filteredGroupVersions = append(filteredGroupVersions, resourceType)
	}
	return filteredGroupVersions
}
