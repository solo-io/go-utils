package types

import (
	"context"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

// This interface implements the Helm CLI actions. The implementation relies on the Helm 3 libraries.
type HelmClient interface {
	// Prepare an installation object that can then be .Run() with a chart object
	NewInstall(namespace, releaseName string, dryRun bool) (HelmInstaller, *cli.EnvSettings, error)

	// Prepare an un-installation object that can then be .Run() with a release name
	NewUninstall(namespace string) (HelmUninstaller, error)

	// List the already-existing releases in the given namespace
	ReleaseList(namespace string) (ReleaseListRunner, error)

	// Returns the Helm chart archive located at the given URI (can be either an http(s) address or a file path)
	DownloadChart(chartArchiveUri string) (*chart.Chart, error)

	// Returns true if the release with the given name exists in the given namespace
	ReleaseExists(namespace, releaseName string) (releaseExists bool, err error)
}

// an interface around Helm's action.Install struct
type HelmInstaller interface {
	Run(chrt *chart.Chart, vals map[string]interface{}) (*release.Release, error)
}

// an interface around Helm's action.Uninstall struct
type HelmUninstaller interface {
	Run(name string) (*release.UninstallReleaseResponse, error)
}

// an interface around Helm's action.List struct
type ReleaseListRunner interface {
	Run() ([]*release.Release, error)
	SetFilter(filter string)
}

var _ HelmInstaller = &action.Install{}
var _ HelmUninstaller = &action.Uninstall{}

type Installer interface {
	Install(ctx context.Context, installerConfig *InstallerConfig) error
}

type InstallerConfig struct {
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
