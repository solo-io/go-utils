module github.com/solo-io/go-utils

go 1.13

require (
	cloud.google.com/go/pubsub v1.3.1
	cloud.google.com/go/storage v1.10.0
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/Netflix/go-expect v0.0.0-20180928190340-9d1f4485533b
	github.com/avast/retry-go v2.2.0+incompatible
	github.com/fgrosse/zaptest v1.1.0
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-git/go-git/v5 v5.4.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/mock v1.5.0
	github.com/golang/protobuf v1.4.3
	github.com/google/go-github/v32 v32.0.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hinshun/vt10x v0.0.0-20180809195222-d55458df857c
	github.com/imroc/req v0.3.0
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
	github.com/yuin/goldmark v1.3.3
	go.opencensus.io v0.23.0
	go.uber.org/zap v1.10.0
	goji.io v2.0.2+incompatible
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a // indirect
	golang.org/x/mod v0.4.1
	golang.org/x/oauth2 v0.0.0-20210313182246-cd4f82c27b84
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/api v0.42.0
	google.golang.org/grpc v1.36.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/AlecAivazis/survey.v1 v1.8.2
	gopkg.in/src-d/go-git.v4 v4.10.0
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.0.5

	// pin to the jwt-go fork to fix CVE.
	// using the pseudo version of github.com/form3tech-oss/jwt-go@v3.2.3 instead of the version directly,
	// to avoid error about it being used for two different module paths
	github.com/dgrijalva/jwt-go => github.com/form3tech-oss/jwt-go v0.0.0-20210511163231-5b2d2b5f6c34

	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
)
