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
		manifests := inputs.InputGlooManifests("myns")
		resources, err := manifests.ResourceList()
		Expect(err).NotTo(HaveOccurred())
		Expect(resources).To(HaveLen(135))
	})
	It("converts resources to a manifest without erroring", func() {
		manifests := inputs.InputGlooManifests("myns")
		resources, err := manifests.ResourceList()
		Expect(err).NotTo(HaveOccurred())
		Expect(resources).To(HaveLen(135))

		recombined, err := ManifestsFromResources(resources)
		Expect(err).NotTo(HaveOccurred())
		Expect(recombined).To(HaveLen(1))

		recombined.CombinedString() // test for panic

		// need to compare resources as comments get removed
		// map ordering is not preserved
		recombinedResources, err := recombined.ResourceList()
		Expect(err).NotTo(HaveOccurred())

		Expect(recombinedResources).To(Equal(resources))
	})

	It("handles value overrides correctly", func() {
		values := `
global:
  glooRbac:
    create: false

`
		manifests, err := RenderManifests(
			context.TODO(),
			"https://storage.googleapis.com/solo-public-helm/charts/gloo-1.0.0.tgz",
			values,
			"yella",
			ns,
			"",
		)
		Expect(err).NotTo(HaveOccurred())

		for _, man := range manifests {
			// no security crds
			Expect(strings.ToLower(man.Content)).NotTo(ContainSubstring("rbac"))
		}

	})

	Context("load from github", func() {
		It("works for a single chart", func() {
			ref := GithubChartRef{
				Ref:            "c00d740c83ec5b5439bf2f58fee110a546af7177",
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
