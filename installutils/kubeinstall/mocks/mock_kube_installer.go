package mocks

import (
	"context"

	"github.com/solo-io/go-utils/installutils/kubeinstall"
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

func (i *MockKubeInstaller) ReconcileResources(ctx context.Context, params kubeinstall.ReconcileParams) error {
	i.ReconcileCalledWith = ReconcileParams{params.InstallNamespace, params.Resources, params.OwnerLabels}
	return i.ReturnErr
}

func (i *MockKubeInstaller) PurgeResources(ctx context.Context, withLabels map[string]string) error {
	i.PurgeCalledWith = PurgeParams{withLabels}
	return i.ReturnErr
}

func (i *MockKubeInstaller) ListAllResources(ctx context.Context) kuberesource.UnstructuredResources {
	return i.ReconcileCalledWith.Resources
}
