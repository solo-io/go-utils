This package includes: 
- A small client interface for loading and storing kubernetes config maps, with a kube and mock implementation. 
- A small wrapper client interface for loading and storing a proto struct from the contents of a config map. 

Example usage: 

In the service mesh hub, we use a config map to store the registries, and have a default config when the config map 
is not found. We can create a light wrapper around the config-to-proto client that is strongly typed for the specific
proto message that the hub uses for it's config struct. 

```go
const (
	ApiserverConfigKey     = "config.yaml"
	ApiserverConfigMapName = "apiserver-config"
	DefaultLogLevel        = v1.LogLevel_INFO_LEVEL
)

func GetDefaultRegistryGithubLocation() *hubv1.GithubRepositoryLocation {
	return &hubv1.GithubRepositoryLocation{
		Org:       "solo-io",
		Repo:      "service-mesh-hub",
		Ref:       "master",
		Directory: "extensions/v1",
	}
}

func GetDefaultApiserverConfig() *v1.ApiserverConfig {
	defaultGithub := &v1.Registry_Github{
		Github: GetDefaultRegistryGithubLocation(),
	}
	return &v1.ApiserverConfig{
		Registries: []*v1.Registry{
			{
				Name:         "default",
				RegistryType: defaultGithub,
			},
		},
		LogLevel: DefaultLogLevel,
	}
}

type ConfigClient interface {
	GetConfig(ctx context.Context) (*v1.ApiserverConfig, error)
	SetConfig(ctx context.Context, config *v1.ApiserverConfig) error
}

type configClient struct {
	delegate configutils.ConfigClient
}

func NewConfigClient(kube configutils.ConfigMapClient, installNamespace string) ConfigClient {
	delegate := configutils.NewConfigClient(kube, installNamespace, ApiserverConfigMapName, ApiserverConfigKey, GetDefaultApiserverConfig())
	return &configClient{
		delegate: delegate,
	}
}

func (c *configClient) GetConfig(ctx context.Context) (*v1.ApiserverConfig, error) {
	var config v1.ApiserverConfig
	if err := c.delegate.GetConfig(ctx, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *configClient) SetConfig(ctx context.Context, config *v1.ApiserverConfig) error {
	return c.delegate.SetConfig(ctx, config)
}
```