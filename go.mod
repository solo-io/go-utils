module github.com/solo-io/go-utils

go 1.13

require (
	cloud.google.com/go/pubsub v1.3.1
	cloud.google.com/go/storage v1.10.0
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/Netflix/go-expect v0.0.0-20180928190340-9d1f4485533b
	github.com/avast/retry-go v2.2.0+incompatible
	github.com/fgrosse/zaptest v1.1.0
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.2
	github.com/google/go-github/v32 v32.0.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hinshun/vt10x v0.0.0-20180809195222-d55458df857c
	github.com/k0kubun/colorstring v0.0.0-20150214042306-9440f1994b88 // indirect
	github.com/k0kubun/pp v2.3.0+incompatible
	github.com/kr/pty v1.1.5 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/hashstructure v1.0.0
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/palantir/go-baseapp v0.2.3
	github.com/palantir/go-githubapp v0.5.0
	github.com/pelletier/go-toml v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.2.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/procfs v0.0.11 // indirect
	github.com/rotisserie/eris v0.1.1
	github.com/rs/zerolog v1.18.0
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v1.1.3
	go.opencensus.io v0.22.4
	go.uber.org/zap v1.10.0
	goji.io v2.0.2+incompatible
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a // indirect
	golang.org/x/mod v0.3.0
	golang.org/x/oauth2 v0.0.0-20210113205817-d3ed898aa8a3
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	google.golang.org/api v0.30.0
	google.golang.org/grpc v1.31.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/AlecAivazis/survey.v1 v1.8.2
	gopkg.in/src-d/go-git.v4 v4.10.0
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.0.5
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
)
