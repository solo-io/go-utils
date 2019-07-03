package testutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/solo-io/go-utils/testutils"
)

var _ = Describe("GetCurrentFileDirectory", func() {
	It("works", func() {
		f, err := GetCurrentFile()
		Expect(err).NotTo(HaveOccurred())
		Expect(f).To(HaveSuffix("go-utils/testutils/get_current_file_test.go"))
		Expect(f).To(HavePrefix("/"))
	})
})
