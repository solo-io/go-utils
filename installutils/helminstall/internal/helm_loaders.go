package internal

import (
	"io"
	"os"

	"github.com/solo-io/go-utils/installutils/helminstall/types"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/client-go/tools/clientcmd"
)

//go:generate mockgen -destination mocks/mock_helm_loaders.go -source ./helm_loaders.go

type HelmFactories struct {
	ActionConfigFactory ActionConfigFactory
	ActionListFactory   ActionListFactory
	ChartLoader         ChartLoader
}

func NewHelmFactories() HelmFactories {
	actionConfigFactory := NewActionConfigFactory()
	return HelmFactories{
		ActionConfigFactory: actionConfigFactory,
		ActionListFactory:   NewActionListFactory(actionConfigFactory),
		ChartLoader:         NewChartLoader(),
	}
}

type ActionConfigFactory interface {
	NewActionConfigFromFile(kubeConfig, helmKubeContext, namespace string) (*action.Configuration, *cli.EnvSettings, error)
	NewActionConfigFromMemory(config clientcmd.ClientConfig, namespace string) (*action.Configuration, *cli.EnvSettings, error)
}

type actionConfigFactory struct{}

func NewActionConfigFactory() ActionConfigFactory {
	return &actionConfigFactory{}
}

// Returns an action configuration that can be used to create Helm actions and the Helm env settings.
// We currently get the Helm storage driver from the standard HELM_DRIVER env (defaults to 'secret').
func (a *actionConfigFactory) NewActionConfigFromFile(kubeConfig, helmKubeContext, namespace string) (*action.Configuration, *cli.EnvSettings, error) {
	settings := NewCLISettings(kubeConfig, helmKubeContext, namespace)
	actionConfig := new(action.Configuration)

	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), noOpDebugLog); err != nil {
		return nil, nil, err
	}
	return actionConfig, settings, nil
}

func (a *actionConfigFactory) NewActionConfigFromMemory(config clientcmd.ClientConfig, namespace string) (*action.Configuration, *cli.EnvSettings, error) {
	settings := NewCLISettings("", "", namespace)
	restClientGetter := NewInMemoryRESTClientGetter(config)
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(restClientGetter, namespace, os.Getenv("HELM_DRIVER"), noOpDebugLog); err != nil {
		return nil, nil, err
	}
	return actionConfig, settings, nil
}

func noOpDebugLog(_ string, _ ...interface{}) {}

// Returns a ReleaseListRunner
type ActionListFactory interface {
	ReleaseList(actionConfig *action.Configuration, namespace string) types.ReleaseListRunner
}

type actionListFactory struct {
	actionConfigFactory ActionConfigFactory
}

func NewActionListFactory(actionConfigFactory ActionConfigFactory) ActionListFactory {
	return &actionListFactory{actionConfigFactory: actionConfigFactory}
}

func (a *actionListFactory) ReleaseList(actionConfig *action.Configuration, namespace string) types.ReleaseListRunner {
	return &releaseListRunner{
		list: action.NewList(actionConfig),
	}
}

type releaseListRunner struct {
	list *action.List
}

func (h *releaseListRunner) Run() ([]*release.Release, error) {
	return h.list.Run()
}

func (h *releaseListRunner) SetFilter(filter string) {
	h.list.Filter = filter
}

// slim interface on top of loader to avoid unnecessary FS calls
type ChartLoader interface {
	Load(archiveFile io.Reader) (*chart.Chart, error)
}

type chartLoader struct{}

func NewChartLoader() ChartLoader {
	return &chartLoader{}
}

func (c *chartLoader) Load(archiveFile io.Reader) (*chart.Chart, error) {
	return loader.LoadArchive(archiveFile)
}
