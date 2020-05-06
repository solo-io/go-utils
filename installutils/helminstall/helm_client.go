package helminstall

import (
	"github.com/solo-io/go-utils/installutils/helminstall/internal"
	"github.com/solo-io/go-utils/installutils/helminstall/types"
	"k8s.io/client-go/tools/clientcmd"
)

// HelmClient factory that accepts kubeconfig as a file.
func DefaultHelmClientFileConfig(kubeConfig, kubeContext string) types.HelmClient {
	return internal.NewHelmClientForFileConfig(
		internal.NewDefaultResourceFetcher(),
		internal.NewHelmFactories(),
		kubeConfig,
		kubeContext,
	)
}

// HelmClient factory that accepts kubeconfig in memory.
func DefaultHelmClientMemoryConfig(config clientcmd.ClientConfig) types.HelmClient {
	return internal.NewHelmClientForMemoryConfig(
		internal.NewDefaultResourceFetcher(),
		internal.NewHelmFactories(),
		config,
	)
}
