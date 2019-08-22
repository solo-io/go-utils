package clusterlock

import (
	"fmt"
	"strings"

	"github.com/hashicorp/consul/api"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// a ClusterLock is only ever composed of either
// a kubernetes configmap, or a consul key
type ClusterLock struct {
	Name    string
	OwnerID string
	Timeout string

	// only for kube
	ResourceVersion string
}

func (l *ClusterLock) Empty() bool {
	return l.OwnerID == "" && l.Timeout == ""
}

func (l *ClusterLock) Set(ownerId, timeout string) {
	l.OwnerID = ownerId
	l.Timeout = timeout
}

func (l ClusterLock) ConfigMap(namespace string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      l.Name,
			Annotations: map[string]string{
				LockAnnotationKey:        l.OwnerID,
				LockTimeoutAnnotationKey: l.Timeout,
			},
			ResourceVersion: l.ResourceVersion,
		},
	}
}
func CLFromConfigMap(cm *v1.ConfigMap) *ClusterLock {
	var ownerId, timeout string
	if cm.Annotations != nil {
		ownerId = cm.Annotations[LockAnnotationKey]
		timeout = cm.Annotations[LockTimeoutAnnotationKey]
	}

	return &ClusterLock{
		Name:            cm.Name,
		OwnerID:         ownerId,
		Timeout:         timeout,
		ResourceVersion: cm.ResourceVersion,
	}
}

var separator = "@"

func toData(ownerId, timeout string) []byte {
	return []byte(fmt.Sprintf("%v%v%v", ownerId, separator, timeout))
}

func fromData(data []byte) (string, string) {
	parts := strings.Split(string(data), separator)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func (l ClusterLock) KVPair(keyPrefix string) *api.KVPair {
	return &api.KVPair{
		Key:   keyPrefix + l.Name,
		Value: toData(l.OwnerID, l.Timeout),
	}
}

func CLFromKVPair(keyPrefix string, kvp *api.KVPair) *ClusterLock {
	ownerId, timeout := fromData(kvp.Value)

	return &ClusterLock{
		Name:    strings.TrimPrefix(kvp.Key, keyPrefix),
		OwnerID: ownerId,
		Timeout: timeout,
	}
}

type ClusterLockClient interface {
	Create(cl *ClusterLock) (*ClusterLock, error)
	Update(cl *ClusterLock) (*ClusterLock, error)
	Get(name string) (*ClusterLock, error)
	Delete(name string) error
}

type KubeClusterLockClient struct {
	clientset kubernetes.Interface
	namespace string
}

func (c *KubeClusterLockClient) Create(cl *ClusterLock) (*ClusterLock, error) {
	cm, err := c.clientset.CoreV1().ConfigMaps(c.namespace).Create(cl.ConfigMap(c.namespace))
	if err != nil {
		return nil, err
	}
	return CLFromConfigMap(cm), nil
}

func (c *KubeClusterLockClient) Update(cl *ClusterLock) (*ClusterLock, error) {
	cm, err := c.clientset.CoreV1().ConfigMaps(c.namespace).Update(cl.ConfigMap(c.namespace))
	if err != nil {
		return nil, err
	}
	return CLFromConfigMap(cm), nil
}

func (c *KubeClusterLockClient) Get(name string) (*ClusterLock, error) {
	cm, err := c.clientset.CoreV1().ConfigMaps(c.namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return CLFromConfigMap(cm), nil
}

func (c *KubeClusterLockClient) Delete(name string) error {
	return c.clientset.CoreV1().ConfigMaps(c.namespace).Delete(name, &metav1.DeleteOptions{})
}

func NotFoundError(name string) error {
	return errors.NewNotFound(v1.Resource("configmap"), name)
}

type ConsulClusterLockClient struct {
	client    *api.Client
	keyPrefix string
}

func (c *ConsulClusterLockClient) Create(cl *ClusterLock) (*ClusterLock, error) {
	kvp := cl.KVPair(c.keyPrefix)
	_, err := c.client.KV().Put(kvp, nil)
	if err != nil {
		return nil, err
	}
	return cl, nil
}

func (c *ConsulClusterLockClient) Update(cl *ClusterLock) (*ClusterLock, error) {
	return c.Create(cl)
}

func (c *ConsulClusterLockClient) Get(name string) (*ClusterLock, error) {
	kvp, _, err := c.client.KV().Get(c.keyPrefix+name, nil)
	if err != nil {
		return nil, err
	}
	if kvp == nil {
		return nil, NotFoundError(name)
	}
	return CLFromKVPair(c.keyPrefix, kvp), nil
}

func (c *ConsulClusterLockClient) Delete(name string) error {
	_, err := c.client.KV().Delete(c.keyPrefix+name, nil)
	return err
}
