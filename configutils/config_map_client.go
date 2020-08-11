package configutils

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	kubeerr "k8s.io/apimachinery/pkg/api/errors"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ConfigMapClient interface {
	GetConfigMap(ctx context.Context, namespace string, name string) (*v1.ConfigMap, error)
	SetConfigMap(ctx context.Context, config *v1.ConfigMap) error
}

type KubeConfigMapClient struct {
	client kubernetes.Interface
}

func NewConfigMapClient(client kubernetes.Interface) ConfigMapClient {
	return &KubeConfigMapClient{
		client: client,
	}
}

func (c *KubeConfigMapClient) GetConfigMap(ctx context.Context, namespace string, configMapName string) (*v1.ConfigMap, error) {
	contextutils.LoggerFrom(ctx).Debugw("Getting config map from Kubernetes",
		zap.String("namespace", namespace),
		zap.String("name", configMapName))
	configMap, err := c.client.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, kubemeta.GetOptions{})
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Could not get config map",
			zap.Error(err),
			zap.String("name", configMapName),
			zap.String("namespace", namespace))
		return nil, err
	}
	return configMap, nil
}

func (c *KubeConfigMapClient) SetConfigMap(ctx context.Context, config *v1.ConfigMap) error {
	contextutils.LoggerFrom(ctx).Debugw("Setting config map in Kubernetes",
		zap.String("namespace", config.Namespace),
		zap.String("name", config.Name))
	_, err := c.client.CoreV1().ConfigMaps(config.Namespace).Update(ctx, config, kubemeta.UpdateOptions{})
	if err != nil {
		if !kubeerr.IsNotFound(err) {
			contextutils.LoggerFrom(ctx).Errorw("Could not update config map",
				zap.Error(err),
				zap.String("name", config.Name),
				zap.String("namespace", config.Namespace),
				zap.Any("configMap", config))
			return err
		}
		_, err := c.client.CoreV1().ConfigMaps(config.Namespace).Create(ctx, config, kubemeta.CreateOptions{})
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("Config map not found, but error creating it",
				zap.Error(err),
				zap.String("name", config.Name),
				zap.String("namespace", config.Namespace),
				zap.Any("configMap", config))
			return err
		}
	}
	return nil
}
