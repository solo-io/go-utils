package internal

import (
	"os"

	"github.com/solo-io/go-utils/installutils/helminstall/types"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	TempChartFilePermissions = os.FileMode(0644)
	TempChartPrefix          = "temp-helm-chart"
	helmNamespaceEnvVar      = "HELM_NAMESPACE"
	helmKubeContextEnvVar    = "HELM_KUBECONTEXT"
)

type helmClient struct {
	resourceFetcher ResourceFetcher
	helmLoaders     HelmFactories
	kubeConfig      string
	kubeContext     string
	config          clientcmd.ClientConfig
}

// Accepts kubeconfig from memory.
func NewHelmClientForMemoryConfig(
	resourceFetcher ResourceFetcher,
	helmLoaders HelmFactories,
	config clientcmd.ClientConfig,
) types.HelmClient {
	return &helmClient{
		resourceFetcher: resourceFetcher,
		helmLoaders:     helmLoaders,
		config:          config,
	}
}

// Accepts kubeconfig persisted on disk.
func NewHelmClientForFileConfig(
	resourceFetcher ResourceFetcher,
	helmLoaders HelmFactories,
	kubeConfig, kubeContext string,
) types.HelmClient {
	return &helmClient{
		resourceFetcher: resourceFetcher,
		helmLoaders:     helmLoaders,
		kubeConfig:      kubeConfig,
		kubeContext:     kubeContext,
	}
}

func (d *helmClient) NewInstall(namespace, releaseName string, dryRun bool) (types.HelmInstaller, *cli.EnvSettings, error) {
	actionConfig, settings, err := d.buildActionConfigAndSettings(namespace)
	if err != nil {
		return nil, nil, err
	}
	client := action.NewInstall(actionConfig)
	client.ReleaseName = releaseName
	client.Namespace = namespace
	client.DryRun = dryRun
	// If this is a dry run, we don't want to query the API server.
	// In the future we can make this configurable to emulate the `helm template --validate` behavior.
	client.ClientOnly = dryRun
	return client, settings, nil
}

func (d *helmClient) NewUninstall(namespace string) (types.HelmUninstaller, error) {
	actionConfig, _, err := d.buildActionConfigAndSettings(namespace)
	if err != nil {
		return nil, err
	}
	return action.NewUninstall(actionConfig), nil
}

func (d *helmClient) DownloadChart(chartArchiveUri string) (*chart.Chart, error) {
	chartFileReader, err := d.resourceFetcher.GetResource(chartArchiveUri)
	if err != nil {
		return nil, err
	}
	defer func() { chartFileReader.Close() }()
	chartObj, err := d.helmLoaders.ChartLoader.Load(chartFileReader)
	if err != nil {
		return nil, err
	}
	return chartObj, nil
}

func (d *helmClient) ReleaseList(namespace string) (types.ReleaseListRunner, error) {
	actionConfig, _, err := d.buildActionConfigAndSettings(namespace)
	if err != nil {
		return nil, err
	}
	return d.helmLoaders.ActionListFactory.ReleaseList(actionConfig, namespace), nil
}

func (d *helmClient) ReleaseExists(namespace, releaseName string) (bool, error) {
	list, err := d.ReleaseList(namespace)
	if err != nil {
		return false, err
	}
	list.SetFilter(releaseName)

	releases, err := list.Run()
	if err != nil {
		return false, err
	}

	releaseExists := false
	for _, r := range releases {
		releaseExists = releaseExists || r.Name == releaseName
	}

	return releaseExists, nil
}

func (d *helmClient) buildActionConfigAndSettings(namespace string) (actionConfig *action.Configuration, settings *cli.EnvSettings, err error) {
	if d.config != nil {
		actionConfig, settings, err = d.helmLoaders.ActionConfigFactory.NewActionConfigFromMemory(d.config, namespace)
	} else {
		actionConfig, settings, err = d.helmLoaders.ActionConfigFactory.NewActionConfigFromFile(d.kubeConfig, d.kubeContext, namespace)
	}
	if err != nil {
		return nil, nil, err
	}
	return actionConfig, settings, nil
}

// Build a Helm EnvSettings struct
// basically, abstracted cli.New() into our own function call because of the weirdness described in the big comment below
// also configure the Helm client with the Kube config/context of the cluster to perform installation on
func NewCLISettings(kubeConfig, kubeContext, namespace string) *cli.EnvSettings {
	// The installation namespace is expressed as a "config override" in the Helm internals
	// It's normally set by the --namespace flag when invoking the Helm binary, which ends up
	// setting a non-exported field in the Helm settings struct (https://github.com/helm/helm/blob/v3.0.1/pkg/cli/environment.go#L77)
	// However, we are not invoking the Helm binary, so that field doesn't get set. It is left as "", which means
	// that any resources that are non-namespaced (at the time of writing, some of Prometheus's resources do not
	// have a namespace attached to them but they probably should) wind up in the default namespace from YOUR
	// kube config. To get around this, we temporarily set an env var before the Helm settings are initialized
	// so that the proper namespace override is piped through. (https://github.com/helm/helm/blob/v3.0.1/pkg/cli/environment.go#L64)
	if os.Getenv(helmNamespaceEnvVar) == "" {
		os.Setenv(helmNamespaceEnvVar, namespace)
		defer os.Setenv(helmNamespaceEnvVar, "")
	}
	if os.Getenv(helmKubeContextEnvVar) == "" {
		os.Setenv(helmKubeContextEnvVar, kubeContext)
		defer os.Setenv(helmNamespaceEnvVar, "")
	}
	settings := cli.New()

	settings.KubeConfig = kubeConfig
	return settings
}
