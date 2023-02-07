package githubutils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/githubutils"
)

var _ = Describe("repo utils", func() {
	It("Gets commit for tag from public repo (no auth)", func() {
		sha, err := githubutils.GetCommitForTag("solo-io", "go-utils", "v0.10.0", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(sha).To(Equal("075c4b07de59a9b7b20475b7eab02dcdc6a62d64"))
	})

	It("gets commit for tag from private repo (auth by GITHUB_TOKEN env var)", func() {
		sha, err := githubutils.GetCommitForTag("solo-io", "solo-projects", "v1.6.0", true)
		Expect(err).NotTo(HaveOccurred())
		Expect(sha).To(Equal("7110f444371f7ea7b18ed4380438709492d02bb8"))

	})
})
