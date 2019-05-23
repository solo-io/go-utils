package configutils

import (
	"context"

	"github.com/gogo/protobuf/proto"

	"github.com/solo-io/go-utils/protoutils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pkg/errors"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
	kubeerr "k8s.io/apimachinery/pkg/api/errors"
)

var (
	ErrorMarshallingConfig = func(err error) error {
		return errors.Wrapf(err, "could not marshal config to string")
	}
	ErrorUnmarshallingConfig = func(err error) error {
		return errors.Wrapf(err, "could not unmarshal string to config")
	}
	ErrorLoadingExistingConfig = func(err error) error {
		return errors.Wrapf(err, "could not load existing config")
	}
	ErrorUpdatingConfig = func(err error) error {
		return errors.Wrapf(err, "could not update config")
	}
	ErrorSettingDefaultConfig = func(err error) error {
		return errors.Wrapf(err, "could not get config, and could not set config to default")
	}
)

type ConfigClient interface {
	GetConfig(ctx context.Context, config proto.Message) error
	SetConfig(ctx context.Context, config proto.Message) error
}

type configClient struct {
	kube               ConfigMapClient
	configMapNamespace string
	configMapName      string
	configKey          string
	defaultConfig      proto.Message
}

func NewConfigClient(kube ConfigMapClient, configMapNamespace, configMapName, configKey string, defaultConfig proto.Message) ConfigClient {
	return &configClient{
		kube:               kube,
		configMapNamespace: configMapNamespace,
		configMapName:      configMapName,
		defaultConfig:      defaultConfig,
		configKey:          configKey,
	}
}

func (c *configClient) getConfigMap(ctx context.Context, config proto.Message) (*corev1.ConfigMap, error) {
	configString, err := WriteConfigToString(ctx, config)
	if err != nil {
		return nil, err
	}
	data := map[string]string{c.configKey: configString}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.configMapNamespace,
			Name:      c.configMapName,
		},
		Data: data,
	}, nil
}

func ReadConfig(ctx context.Context, value string, config proto.Message) error {
	if err := protoutils.UnmarshalYaml([]byte(value), config); err != nil {
		wrapped := ErrorUnmarshallingConfig(err)
		contextutils.LoggerFrom(ctx).Errorw(wrapped.Error(),
			zap.Error(err),
			zap.Any("value", value))
		return wrapped
	}
	return nil
}

func WriteConfigToString(ctx context.Context, config proto.Message) (string, error) {
	bytes, err := protoutils.MarshalBytes(config)
	if err != nil {
		wrapped := ErrorMarshallingConfig(err)
		contextutils.LoggerFrom(ctx).Errorw(wrapped.Error(),
			zap.Error(err),
			zap.Any("config", config))
		return "", wrapped
	}
	return string(bytes), nil
}

func (c *configClient) GetConfig(ctx context.Context, config proto.Message) error {
	contextutils.LoggerFrom(ctx).Debugw("Loading config",
		zap.String("configMapName", c.configMapName),
		zap.String("configMapNamespace", c.configMapNamespace),
		zap.String("configKey", c.configKey))
	loaded, err := c.kube.GetConfigMap(ctx, c.configMapNamespace, c.configMapName)
	if err != nil && !kubeerr.IsNotFound(err) {
		return ErrorLoadingExistingConfig(err)
	} else if err != nil {
		defaultConfig := c.defaultConfig
		defaultConfigMap, marshalErr := c.getConfigMap(ctx, defaultConfig)
		if marshalErr != nil {
			return marshalErr
		}
		setDefaultErr := c.kube.SetConfigMap(ctx, defaultConfigMap)
		if setDefaultErr != nil {
			return ErrorSettingDefaultConfig(setDefaultErr)
		}
		loaded = defaultConfigMap
	}
	return ReadConfig(ctx, loaded.Data[c.configKey], config)
}

func (c *configClient) SetConfig(ctx context.Context, config proto.Message) error {
	contextutils.LoggerFrom(ctx).Infow("Storing config",
		zap.Any("config", config),
		zap.String("configMapNamespace", c.configMapNamespace),
		zap.String("configMapName", c.configMapName),
		zap.String("configKey", c.configKey))
	configString, err := WriteConfigToString(ctx, config)
	if err != nil {
		return err
	}
	data := map[string]string{c.configKey: configString}
	var configMap *corev1.ConfigMap
	// The config map should always exist, but if it was deleted then this will log an error and return a nil,
	// which we can handle gracefully
	loaded, _ := c.kube.GetConfigMap(ctx, c.configMapNamespace, c.configMapName)
	if loaded == nil {
		configMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: c.configMapNamespace,
				Name:      c.configMapName,
			},
			Data: data,
		}
	} else {
		configMap = loaded
		configMap.Data = data
	}
	err = c.kube.SetConfigMap(ctx, configMap)
	if err != nil {
		return ErrorUpdatingConfig(err)
	}
	return nil
}
