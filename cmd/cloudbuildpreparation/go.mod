module github.com/solo-io/go-utils/cmd/cloudbuildpreparation

go 1.12

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.0.5
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

	//kube 1.16
	k8s.io/api => k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
)

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/boltdb/bolt v1.3.1 // indirect
	github.com/golang/dep v0.5.4 // indirect
	github.com/google/go-github v17.0.0+incompatible
	github.com/jmank88/nuts v0.4.0 // indirect
	github.com/nightlyone/lockfile v0.0.0-20180618180623-0ad87eef1443 // indirect
	github.com/sdboyer/constext v0.0.0-20170321163424-836a14457353 // indirect
	github.com/solo-io/go-utils v0.11.0
	github.com/spf13/afero v1.2.2
	go.uber.org/zap v1.13.0
)
