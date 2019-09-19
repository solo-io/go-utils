package git_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/versionutils/git"
)

var _ = Describe("Git Versions", func() {

	It("works as expected", func() {
		refInfo, err := git.GetGitRefInfo("./../..")
		Expect(err).NotTo(HaveOccurred())
		Expect(refInfo).NotTo(BeNil())
		// Git commit sha is always 40 characters long
		Expect(refInfo.Hash).To(HaveLen(40))
		// Our tags always start with"v"
		Expect(refInfo.Tag).To(HavePrefix("v"))
		// Nothing reliable to test the branch with, unfortunately
	})
})
