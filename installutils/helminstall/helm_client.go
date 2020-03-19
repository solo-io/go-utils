package helminstall

import (
	"github.com/solo-io/go-utils/installutils/helminstall/internal"
	"github.com/solo-io/go-utils/installutils/helminstall/types"
	"github.com/spf13/afero"
)

// a HelmClient that talks to the kube api server and creates resources
func DefaultHelmClient() types.HelmClient {
	return &internal.DefaultHelmClient{
		Fs:              internal.NewFs(afero.NewOsFs()),
		ResourceFetcher: internal.NewDefaultResourceFetcher(),
		HelmLoaders:     internal.NewHelmFactories(),
	}
}
