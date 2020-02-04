package helminstall

import (
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

//go:generate mockgen -destination mocks/mock_helm_loaders.go -source ./helm_loaders.go

type HelmLoaders struct {
	ActionConfigLoader ActionConfigLoader
	ActionListLoader   ActionListLoader
	ChartLoader        ChartLoader
}

func NewHelmLoaders() HelmLoaders {
	return HelmLoaders{
		ActionConfigLoader: NewActionConfigLoader(),
		ActionListLoader:   NewActionListLoader(),
		ChartLoader:        NewChartLoader(),
	}
}

type ActionConfigLoader interface {
	NewActionConfig(namespace string) (*action.Configuration, *cli.EnvSettings, error)
}

type actionConfigLoader struct{}

func NewActionConfigLoader() ActionConfigLoader {
	return &actionConfigLoader{}
}

// Returns an action configuration that can be used to create Helm actions and the Helm env settings.
// We currently get the Helm storage driver from the standard HELM_DRIVER env (defaults to 'secret').
func (a *actionConfigLoader) NewActionConfig(namespace string) (*action.Configuration, *cli.EnvSettings, error) {
	settings := NewCLISettings(namespace)
	actionConfig := new(action.Configuration)

	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), noOpDebugLog); err != nil {
		return nil, nil, err
	}
	return actionConfig, settings, nil
}

func noOpDebugLog(_ string, _ ...interface{}) {}

// Returns a ReleaseListRunner
type ActionListLoader interface {
	ReleaseList(helmActionConfigLoader ActionConfigLoader, namespace string) (ReleaseListRunner, error)
}

type actionListLoader struct{}

func NewActionListLoader() ActionListLoader {
	return &actionListLoader{}
}

func (h *actionListLoader) ReleaseList(actionConfigLoader ActionConfigLoader, namespace string) (ReleaseListRunner, error) {
	actionConfig, _, err := actionConfigLoader.NewActionConfig(namespace)
	if err != nil {
		return nil, err
	}
	return &releaseListRunner{
		list: action.NewList(actionConfig),
	}, nil
}

// an interface around Helm's action.List struct
type ReleaseListRunner interface {
	Run() ([]*release.Release, error)
	SetFilter(filter string)
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
	Load(name string) (*chart.Chart, error)
}

type chartLoader struct{}

func NewChartLoader() ChartLoader {
	return &chartLoader{}
}

func (c *chartLoader) Load(name string) (*chart.Chart, error) {
	return loader.Load(name)
}
