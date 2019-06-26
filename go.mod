module github.com/solo-io/go-utils

go 1.12

require (
	cloud.google.com/go v0.40.0
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Netflix/go-expect v0.0.0-20180928190340-9d1f4485533b
	github.com/Sirupsen/logrus v1.4.2 // indirect
	github.com/avast/retry-go v2.2.0+incompatible
	github.com/bluekeyes/hatpear v0.0.0-20180714193905-ffb42d5bb417 // indirect
	github.com/bradleyfalzon/ghinstallation v0.1.2 // indirect
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/evanphx/json-patch v4.1.0+incompatible
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/fgrosse/zaptest v1.1.0
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.2.1
	github.com/golang/mock v1.3.1
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/goph/emperror v0.17.1
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.5.0 // indirect
	github.com/hinshun/vt10x v0.0.0-20180809195222-d55458df857c
	github.com/jonboulle/clockwork v0.1.0 // indirect
	github.com/json-iterator/go v1.1.6 // indirect
	github.com/k0kubun/pp v2.3.0+incompatible
	github.com/mitchellh/go-homedir v1.0.0
	github.com/mitchellh/hashstructure v1.0.0
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/palantir/go-baseapp v0.0.0-20181120182528-91001a7c54f3
	github.com/palantir/go-githubapp v0.0.0-20181023164003-a3a32b23c2f7
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pelletier/go-toml v1.2.0
	github.com/pkg/errors v0.8.1
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a // indirect
	github.com/rs/zerolog v1.14.3
	github.com/shurcooL/githubv4 v0.0.0-20190625031733-ee671ab25ff0 // indirect
	github.com/shurcooL/graphql v0.0.0-20181231061246-d48a9a75455f // indirect
	github.com/soheilhy/cmux v0.1.4 // indirect
	github.com/solo-io/build v0.0.5
	github.com/solo-io/solo-kit v0.6.3
	github.com/spf13/afero v1.2.1
	github.com/spf13/cobra v0.0.3
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/ugorji/go v1.1.5-pre // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	github.com/xlab/handysort v0.0.0-20150421192137-fb3537ed64a1 // indirect
	go.etcd.io/bbolt v1.3.3 // indirect
	go.opencensus.io v0.22.0
	go.uber.org/zap v1.9.1
	goji.io v2.0.2+incompatible
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/xerrors v0.0.0-20190513163551-3ee3066db522
	google.golang.org/api v0.6.0
	google.golang.org/grpc v1.20.1
	gopkg.in/AlecAivazis/survey.v1 v1.8.2
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	gopkg.in/src-d/go-git.v4 v4.10.0
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20181221193117-173ce66c1e39+incompatible
	k8s.io/apiextensions-apiserver v0.0.0-20190111034747-7d26de67f177+incompatible
	k8s.io/apimachinery v0.0.0-20190104073114-849b284f3b75+incompatible
	k8s.io/client-go v10.0.0+incompatible
	k8s.io/helm v2.13.0+incompatible
	k8s.io/klog v0.3.3 // indirect
	k8s.io/kubernetes v1.13.2
	sigs.k8s.io/controller-runtime v0.1.10
	sigs.k8s.io/yaml v1.1.0
)

replace (
	github.com/Sirupsen/logrus v1.0.5 => github.com/sirupsen/logrus v1.0.5
	github.com/Sirupsen/logrus v1.3.0 => github.com/Sirupsen/logrus v1.0.6
	github.com/Sirupsen/logrus v1.4.2 => github.com/sirupsen/logrus v1.0.6
)
