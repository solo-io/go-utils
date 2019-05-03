package installutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/installutils"
)

var _ = Describe("Manifests", func() {
	It("works", func() {
		manifests, err := installutils.GetManifestsFromRemoteTar("https://github.com/XiaoMi/naftis/releases/download/0.1.4-rc6/manifest.tar.gz")
		Expect(err).NotTo(HaveOccurred())
		Expect(len(manifests)).To(BeEquivalentTo(3))
	})
})
