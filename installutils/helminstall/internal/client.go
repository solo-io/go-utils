package internal

import (
	"io/ioutil"
	"os"

	"github.com/spf13/afero"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

//go:generate mockgen -destination mocks/mock_helm_client.go -source ./client.go

const (
	TempChartFilePermissions = os.FileMode(0644)
	TempChartPrefix          = "temp-helm-chart"
	helmNamespaceEnvVar      = "HELM_NAMESPACE"
	helmKubeContextEnvVar    = "HELM_KUBECONTEXT"
)

// an interface around Helm's action.Install struct
type HelmInstaller interface {
	Run(chrt *chart.Chart, vals map[string]interface{}) (*release.Release, error)
}

// an interface around Helm's action.Uninstall struct
type HelmUninstaller interface {
	Run(name string) (*release.UninstallReleaseResponse, error)
}

var _ HelmInstaller = &action.Install{}
var _ HelmUninstaller = &action.Uninstall{}

// interface around needed afero functions
type FsHelper interface {
	NewTempFile(dir, prefix string) (f afero.File, err error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
	RemoveAll(path string) error
}

type tempFile struct {
	fs afero.Fs
}

func NewFs(fs afero.Fs) FsHelper {
	return &tempFile{fs: fs}
}

func (t *tempFile) NewTempFile(dir, prefix string) (f afero.File, err error) {
	return afero.TempFile(t.fs, dir, prefix)
}

func (t *tempFile) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return afero.WriteFile(t.fs, filename, data, perm)
}

func (t *tempFile) RemoveAll(path string) error {
	return t.fs.RemoveAll(path)
}

type DefaultHelmClient struct {
	Fs              FsHelper
	ResourceFetcher ResourceFetcher
	HelmLoaders     HelmFactories
}

func NewDefaultHelmClient(
	fs FsHelper,
	resourceFetcher ResourceFetcher,
	helmLoaders HelmFactories) *DefaultHelmClient {
	return &DefaultHelmClient{
		Fs:              fs,
		ResourceFetcher: resourceFetcher,
		HelmLoaders:     helmLoaders,
	}
}

func (d *DefaultHelmClient) NewInstall(kubeConfig, kubeContext, namespace, releaseName string, dryRun bool) (HelmInstaller, *cli.EnvSettings, error) {
	actionConfig, settings, err := d.HelmLoaders.ActionConfigFactory.NewActionConfig(kubeConfig, kubeContext, namespace)
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

func (d *DefaultHelmClient) NewUninstall(kubeConfig, kubeContext, namespace string) (HelmUninstaller, error) {
	actionConfig, _, err := d.HelmLoaders.ActionConfigFactory.NewActionConfig(kubeConfig, kubeContext, namespace)
	if err != nil {
		return nil, err
	}
	return action.NewUninstall(actionConfig), nil
}

func (d *DefaultHelmClient) DownloadChart(chartArchiveUri string) (*chart.Chart, error) {

	// 1. Get a reader to the chart file (remote URL or local file path)
	chartFileReader, err := d.ResourceFetcher.GetResource(chartArchiveUri)
	if err != nil {
		return nil, err
	}
	defer func() { chartFileReader.Close() }()

	// 2. Write chart to a temporary file
	chartBytes, err := ioutil.ReadAll(chartFileReader)
	if err != nil {
		return nil, err
	}

	chartFile, err := d.Fs.NewTempFile("", TempChartPrefix)
	if err != nil {
		return nil, err
	}
	chartFilePath := chartFile.Name()
	defer func() { d.Fs.RemoveAll(chartFilePath) }()

	if err := d.Fs.WriteFile(chartFilePath, chartBytes, TempChartFilePermissions); err != nil {
		return nil, err
	}

	// 3. Load the chart file
	chartObj, err := d.HelmLoaders.ChartLoader.Load(chartFilePath)
	if err != nil {
		return nil, err
	}

	return chartObj, nil
}

func (d *DefaultHelmClient) ReleaseList(kubeConfig, kubeContext, namespace string) (ReleaseListRunner, error) {
	return d.HelmLoaders.ActionListFactory.ReleaseList(kubeConfig, kubeContext, namespace)
}

func (d *DefaultHelmClient) ReleaseExists(kubeConfig, kubeContext, namespace, releaseName string) (bool, error) {
	list, err := d.ReleaseList(kubeConfig, kubeContext, namespace)
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
