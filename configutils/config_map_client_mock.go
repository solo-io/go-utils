package configutils

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MockConfigMapClient struct {
	Data     map[string]string
	GetError error
	SetError error
}

func (c *MockConfigMapClient) GetConfigMap(ctx context.Context, namespace string, name string) (*v1.ConfigMap, error) {
	if c.GetError != nil {
		return nil, c.GetError
	}
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Data: c.Data,
	}, nil
}

func (c *MockConfigMapClient) SetConfigMap(ctx context.Context, configMap *v1.ConfigMap) error {
	if c.SetError != nil {
		return c.SetError
	}
	return nil
}
