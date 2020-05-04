package helminstall

import (
	"github.com/solo-io/go-utils/installutils/helminstall/internal"
	"github.com/solo-io/go-utils/installutils/helminstall/types"
	"github.com/spf13/afero"
)

// HelmClient factory that accepts kubeconfig as a file.
func DefaultHelmClientFileConfigFactory() types.HelmClientForFileConfigFactory {
	return internal.NewHelmClientForFileConfigFactory(
		internal.NewFs(afero.NewOsFs()),
		internal.NewDefaultResourceFetcher(),
		internal.NewHelmFactories())
}

// HelmClient factory that accepts kubeconfig in memory.
func DefaultHelmClientMemoryConfigFactory() types.HelmClientForMemoryConfigFactory {
	return internal.NewHelmClientForMemoryConfigFactory(
		internal.NewFs(afero.NewOsFs()),
		internal.NewDefaultResourceFetcher(),
		internal.NewHelmFactories())
}
