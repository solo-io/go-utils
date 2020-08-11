package clusterlock

import (
	"context"
	"fmt"
	"strconv"
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

	// either a kube resource version or consul modify index
	ResourceVersion string
}

func (l *ClusterLock) Empty() bool {
	return l.OwnerID == "" && l.Timeout == ""
}

func (l *ClusterLock) Clear() {
	l.Set("", "")
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
	modifyIndex, _ := strconv.Atoi(l.ResourceVersion)
	return &api.KVPair{
		Key:         keyPrefix + l.Name,
		Value:       toData(l.OwnerID, l.Timeout),
		ModifyIndex: uint64(modifyIndex),
	}
}

func CLFromKVPair(keyPrefix string, kvp *api.KVPair) *ClusterLock {
	ownerId, timeout := fromData(kvp.Value)

	return &ClusterLock{
		Name:            strings.TrimPrefix(kvp.Key, keyPrefix),
		OwnerID:         ownerId,
		Timeout:         timeout,
		ResourceVersion: strconv.Itoa(int(kvp.ModifyIndex)),
	}
}

type ClusterLockClient interface {
	Create(ctx context.Context, cl *ClusterLock) (*ClusterLock, error)
	Update(ctx context.Context, cl *ClusterLock) (*ClusterLock, error)
	Get(ctx context.Context, name string) (*ClusterLock, error)
	Delete(ctx context.Context, name string) error
}

type KubeClusterLockClient struct {
	clientset kubernetes.Interface
	namespace string
}

func (c *KubeClusterLockClient) Create(ctx context.Context, cl *ClusterLock) (*ClusterLock, error) {
	cm, err := c.clientset.CoreV1().ConfigMaps(c.namespace).Create(ctx, cl.ConfigMap(c.namespace), metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return CLFromConfigMap(cm), nil
}

func (c *KubeClusterLockClient) Update(ctx context.Context, cl *ClusterLock) (*ClusterLock, error) {
	originalCm, err := c.clientset.CoreV1().ConfigMaps(c.namespace).Get(ctx, cl.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	cm := cl.ConfigMap(c.namespace)
	cm.ResourceVersion = originalCm.ResourceVersion
	cm, err = c.clientset.CoreV1().ConfigMaps(c.namespace).Update(ctx, cl.ConfigMap(c.namespace), metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	return CLFromConfigMap(cm), nil
}

func (c *KubeClusterLockClient) Get(ctx context.Context, name string) (*ClusterLock, error) {
	cm, err := c.clientset.CoreV1().ConfigMaps(c.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return CLFromConfigMap(cm), nil
}

func (c *KubeClusterLockClient) Delete(ctx context.Context, name string) error {
	return c.clientset.CoreV1().ConfigMaps(c.namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func ExistsError(name string) error {
	return errors.NewAlreadyExists(v1.Resource("key-value-pair"), name)
}

func NotFoundError(name string) error {
	return errors.NewNotFound(v1.Resource("key-value-pair"), name)
}

func ConflictError(name string) error {
	return errors.NewConflict(v1.Resource("key-value-pair"), name, fmt.Errorf(""))
}

type ConsulClusterLockClient struct {
	client    *api.Client
	keyPrefix string
}

func (c *ConsulClusterLockClient) Create(ctx context.Context, cl *ClusterLock) (*ClusterLock, error) {
	_, err := c.Get(ctx, cl.Name)
	if err == nil {
		return nil, ExistsError(cl.Name)
	}
	kvp := cl.KVPair(c.keyPrefix)
	success, _, err := c.client.KV().CAS(kvp, nil)
	if err != nil {
		return nil, err
	}
	if !success {
		return nil, ExistsError(cl.Name)
	}
	return c.Get(ctx, cl.Name)
}

func (c *ConsulClusterLockClient) Update(ctx context.Context, cl *ClusterLock) (*ClusterLock, error) {
	_, err := c.Get(ctx, cl.Name)
	if err != nil {
		return nil, err
	}
	kvp := cl.KVPair(c.keyPrefix)
	success, _, err := c.client.KV().CAS(kvp, nil)
	if err != nil {
		return nil, err
	}
	if !success {
		return nil, ConflictError(cl.Name)
	}
	return c.Get(ctx, cl.Name)
}

func (c *ConsulClusterLockClient) put(cl *ClusterLock) (*ClusterLock, error) {
	kvp := cl.KVPair(c.keyPrefix)
	success, _, err := c.client.KV().CAS(kvp, nil)
	if err != nil {
		return nil, err
	}
	if !success {
		return nil, ConflictError(cl.Name)
	}
	return cl, nil
}

func (c *ConsulClusterLockClient) Get(_ context.Context, name string) (*ClusterLock, error) {
	kvp, _, err := c.client.KV().Get(c.keyPrefix+name, nil)
	if err != nil {
		return nil, err
	}
	if kvp == nil {
		return nil, NotFoundError(name)
	}
	return CLFromKVPair(c.keyPrefix, kvp), nil
}

func (c *ConsulClusterLockClient) Delete(_ context.Context, name string) error {
	_, err := c.client.KV().Delete(c.keyPrefix+name, nil)
	return err
}
