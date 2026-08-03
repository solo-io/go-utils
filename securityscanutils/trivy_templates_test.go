package securityscanutils_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/securityscanutils"
)

var _ = Describe("Trivy Templates", func() {
	It("creates a .tpl file containing the requested template", func() {
		templateFile, err := GetTemplateFile(MarkdownTrivyTemplate)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(os.Remove, templateFile)

		Expect(filepath.Ext(templateFile)).To(Equal(".tpl"))
		contents, err := os.ReadFile(templateFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(Equal(MarkdownTrivyTemplate))
	})
})
