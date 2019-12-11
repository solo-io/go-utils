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
	github.com/solo-io/go-utils v0.11.0 // indirect
	go.uber.org/zap v1.13.0 // indirect
)
