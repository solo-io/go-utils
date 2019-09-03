package versionutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/versionutils"
)

var _ = Describe("Simple Toml Parser", func() {

	var _ = Context("parsing image version file", func() {
		It("works", func() {
			fileContent := `
[[image]]
  id = "gloo-ui"
  name = "gloo-ui"
  repo = "quay.io/solo-io"
  tag = "v1.2.3"
[[image]]
  id = "envoy-alpha"
  name = "envoy"
  repo = "quay.io/solo-io"
  tag = "v1.2.3alpha"
[[image]]
  id = "envoy-beta"
  name = "envoy"
  repo = "quay.io/solo-io"
  tag = "v1.2.3beta"
[[image]]
  id = "missing-image-name"
  repo = "quay.io/solo-io"
  tag = "v1.2.3beta"
[[image]]
  id = "missing-image-repo"
  name = "envoy"
  tag = "v1.2.3beta"
[[image]]
  id = "missing-image-tag"
  name = "envoy"
  repo = "quay.io/solo-io"
`
			stp, err := versionutils.NewSimpleTomlParserFromString(fileContent)
			Expect(err).NotTo(HaveOccurred())

			glooUiVersion, err := stp.GetImageVersionFromToml("gloo-ui")
			Expect(err).NotTo(HaveOccurred())
			Expect(glooUiVersion).To(Equal("quay.io/solo-io/gloo-ui:v1.2.3"))

			envoyAlphaVersion, err := stp.GetImageVersionFromToml("envoy-alpha")
			Expect(err).NotTo(HaveOccurred())
			Expect(envoyAlphaVersion).To(Equal("quay.io/solo-io/envoy:v1.2.3alpha"))

			envoyBetaVersion, err := stp.GetImageVersionFromToml("envoy-beta")
			Expect(err).NotTo(HaveOccurred())
			Expect(envoyBetaVersion).To(Equal("quay.io/solo-io/envoy:v1.2.3beta"))

			missingImageName, err := stp.GetImageVersionFromToml("missing-image-name")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid image spec: no name specified"))
			Expect(missingImageName).To(Equal(""))

			missingImageRepo, err := stp.GetImageVersionFromToml("missing-image-repo")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid image spec: no repo specified"))
			Expect(missingImageRepo).To(Equal(""))

			missingImageTag, err := stp.GetImageVersionFromToml("missing-image-tag")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid image spec: no tag specified"))
			Expect(missingImageTag).To(Equal(""))

			idNotFound, err := stp.GetImageVersionFromToml("does-not-exist-in-spec")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`root key-value pair "id"="does-not-exist-in-spec" not found`))
			Expect(idNotFound).To(Equal(""))
		})
	})
})
