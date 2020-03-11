package helminstall

import (
	"github.com/solo-io/go-utils/installutils/helminstall/internal"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
)

// This interface implements the Helm CLI actions. The implementation relies on the Helm 3 libraries.
type HelmClient interface {
	// Prepare an installation object that can then be .Run() with a chart object
	NewInstall(kubeConfig, kubeContext, namespace, releaseName string, dryRun bool) (internal.HelmInstaller, *cli.EnvSettings, error)

	// Prepare an un-installation object that can then be .Run() with a release name
	NewUninstall(kubeConfig, kubeContext, namespace string) (internal.HelmUninstaller, error)

	// List the already-existing releases in the given namespace
	ReleaseList(kubeConfig, kubeContext, namespace string) (internal.ReleaseListRunner, error)

	// Returns the Helm chart archive located at the given URI (can be either an http(s) address or a file path)
	DownloadChart(chartArchiveUri string) (*chart.Chart, error)

	// Returns true if the release with the given name exists in the given namespace
	ReleaseExists(kubeConfig, kubeContext, namespace, releaseName string) (releaseExists bool, err error)
}

type Installer interface {
	Install(installerConfig *InstallerConfig) error
}

type InstallerConfig struct {
	// kube config containing the context of cluster to install on
	KubeConfig string
	// kube context of cluster to install on
	KubeContext      string
	DryRun           bool
	CreateNamespace  bool
	Verbose          bool
	InstallNamespace string
	ReleaseName      string
	// the uri to the helm chart, can either be a local file or a valid http/https link
	ReleaseUri  string
	ValuesFiles []string
	ExtraValues map[string]interface{}

	PreInstallMessage  string
	PostInstallMessage string
}
