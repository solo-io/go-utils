module github.com/solo-io/go-utils

go 1.13

require (
	cloud.google.com/go v0.40.0
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/Netflix/go-expect v0.0.0-20180928190340-9d1f4485533b
	github.com/avast/retry-go v2.2.0+incompatible
	github.com/bluekeyes/hatpear v0.0.0-20180714193905-ffb42d5bb417 // indirect
	github.com/bradleyfalzon/ghinstallation v0.1.2 // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/fgrosse/zaptest v1.1.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/mock v1.3.1
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/uuid v1.1.1
	github.com/goph/emperror v0.17.1
	github.com/hashicorp/consul/api v1.1.0
	github.com/hinshun/vt10x v0.0.0-20180809195222-d55458df857c
	github.com/k0kubun/colorstring v0.0.0-20150214042306-9440f1994b88 // indirect
	github.com/k0kubun/pp v2.3.0+incompatible
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/hashstructure v1.0.0
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/palantir/go-baseapp v0.0.0-20190430095958-24408a192334
	github.com/palantir/go-githubapp v0.0.0-20190620114758-91e10a96cd34
	github.com/pelletier/go-toml v1.2.0
	github.com/pkg/errors v0.8.1
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a // indirect
	github.com/rotisserie/eris v0.1.1
	github.com/rs/zerolog v1.14.3
	github.com/shurcooL/githubv4 v0.0.0-20190625031733-ee671ab25ff0 // indirect
	github.com/shurcooL/graphql v0.0.0-20181231061246-d48a9a75455f // indirect
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v0.0.5
	go.opencensus.io v0.21.0
	go.uber.org/zap v1.9.1
	goji.io v2.0.2+incompatible
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/xerrors v0.0.0-20190717185122-a985d3407aa7
	google.golang.org/api v0.6.0
	google.golang.org/grpc v1.24.0
	gopkg.in/AlecAivazis/survey.v1 v1.8.2
	gopkg.in/src-d/go-git.v4 v4.10.0
	gopkg.in/yaml.v2 v2.2.4
	helm.sh/helm/v3 v3.0.0
	k8s.io/api v0.0.0-20191121015604-11707872ac1c
	k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65
	k8s.io/apimachinery v0.0.0-20191121015412-41065c7a8c2a
	k8s.io/cli-runtime v0.0.0-20191016114015-74ad18325ed5
	k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/yaml v1.1.0
)

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.0.5
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

	//kube 1.16
	k8s.io/api => k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
)
