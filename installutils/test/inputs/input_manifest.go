package inputs

import (
	"context"

	"github.com/solo-io/go-utils/installutils/helmchart"

	"github.com/onsi/gomega"
)

func InputGlooManifests(ns string) helmchart.Manifests {
	manifests, err := helmchart.RenderManifests(
		context.TODO(),
		"https://storage.googleapis.com/solo-public-helm/charts/gloo-1.0.0.tgz",
		"",
		"yella",
		ns,
		"",
	)
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
	return manifests
}
