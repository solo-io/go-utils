package helmchart_test

import (
	"context"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/installutils/test/inputs"
)

var _ = Describe("Manifests", func() {
	ns := "anything"
	It("converts resources from a manifest without erroring", func() {
		manifests := inputs.InputIstioManifests("myns")
		resources, err := manifests.ResourceList()
		Expect(err).NotTo(HaveOccurred())
		Expect(resources).To(HaveLen(135))
	})

	It("handles value overrides correctly", func() {
		values := `
mixer:
  enabled: false #should not install mixer

`
		manifests, err := RenderManifests(
			context.TODO(),
			"https://s3.amazonaws.com/supergloo.solo.io/istio-1.0.3.tgz",
			values,
			"yella",
			ns,
			"",
		)
		Expect(err).NotTo(HaveOccurred())

		for _, man := range manifests {
			// no security crds
			Expect(man.Content).NotTo(ContainSubstring("policies.authentication.istio.io"))

			// no mixer-policy
			Expect(man.Content).NotTo(ContainSubstring(`apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-policy`))
			// no mixer-telemetry
			Expect(man.Content).NotTo(ContainSubstring(`apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-telemetry`))
		}

	})
	Context("update manifest", func() {
		It("updates the existing deployed resources correctly", func() {
			values := `
mixer:
  enabled: true #should install mixer

`
			manifests, err := RenderManifests(
				context.TODO(),
				"https://s3.amazonaws.com/supergloo.solo.io/istio-1.0.3.tgz",
				values,
				"yella",
				ns,
				"",
			)
			Expect(err).NotTo(HaveOccurred())

			var foundMixerPolicy, foundMixerTelemetry bool
			for _, man := range manifests {
				// yes mixer-policy
				if strings.Contains(man.Content, `apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-policy`) {
					foundMixerPolicy = true
				}
				if strings.Contains(man.Content, `apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-telemetry`) {
					foundMixerTelemetry = true
				}
			}

			Expect(foundMixerPolicy).To(BeTrue())
			Expect(foundMixerTelemetry).To(BeTrue())

		})
	})

	Context("load from github", func() {
		It("works for a single chart", func() {
			ref := GithubChartRef{
				Ref:            "master",
				Repo:           "charts",
				Owner:          "helm",
				ChartDirectory: "stable/envoy",
			}
			manifests, err := RenderManifestsFromGithub(context.TODO(), ref, "", "test", "test", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(len(manifests)).To(BeEquivalentTo(4))
		})

		It("works for a directory of charts", func() {
			ref := GithubChartRef{
				Ref:            "c00d740c83ec5b5439bf2f58fee110a546af7177",
				Repo:           "charts",
				Owner:          "helm",
				ChartDirectory: "stable",
			}
			charts, err := RenderChartsFromGithub(context.TODO(), ref)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(charts)).To(BeEquivalentTo(270))
		})
	})
})
