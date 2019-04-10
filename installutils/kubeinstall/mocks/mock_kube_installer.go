package mocks

import (
	"context"

	"github.com/solo-io/go-utils/installutils/kuberesource"
)

type MockKubeInstaller struct {
	ReconcileCalledWith ReconcileParams
	PurgeCalledWith     PurgeParams
	ReturnErr           error
}

type ReconcileParams struct {
	InstallNamespace string
	Resources        kuberesource.UnstructuredResources
	InstallLabels    map[string]string
}

type PurgeParams struct {
	InstallLabels map[string]string
}

func (i *MockKubeInstaller) ReconcileResources(ctx context.Context, installNamespace string, resources kuberesource.UnstructuredResources, installLabels map[string]string) error {
	i.ReconcileCalledWith = ReconcileParams{installNamespace, resources, installLabels}
	return i.ReturnErr
}

func (i *MockKubeInstaller) PurgeResources(ctx context.Context, withLabels map[string]string) error {
	i.PurgeCalledWith = PurgeParams{withLabels}
	return i.ReturnErr
}

func (i *MockKubeInstaller) ListAllCachedValues(ctx context.Context, labelKey string) []string {
	// check if this key was reconciled
	if i.ReconcileCalledWith.InstallLabels == nil {
		return []string{}
	}
	labelValue := i.ReconcileCalledWith.InstallLabels[labelKey]
	if labelValue == "" {
		return []string{}
	}
	if i.PurgeCalledWith.InstallLabels == nil {
		return []string{labelValue}
	}
	// check if this value was purged
	purgedLabelValue := i.PurgeCalledWith.InstallLabels[labelKey]
	if labelValue == purgedLabelValue {
		return []string{}
	}
	return []string{labelValue}
}
