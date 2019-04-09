package inputs

import (
	"context"

	"github.com/solo-io/go-utils/installutils/helmchart"

	"github.com/onsi/gomega"
)

func InputIstioManifests(ns string) helmchart.Manifests {
	manifests, err := helmchart.RenderManifests(
		context.TODO(),
		"https://s3.amazonaws.com/supergloo.solo.io/istio-1.0.3.tgz",
		"",
		"yella",
		ns,
		"",
	)
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	return manifests
}
