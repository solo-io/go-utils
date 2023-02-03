package testutils_test

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"

    . "github.com/solo-io/go-utils/testutils"
)

var _ = Describe("GetCurrentFileDirectory", func() {
    It("works", func() {
        f, err := GetCurrentFile()
        Expect(err).NotTo(HaveOccurred())
        Expect(f).To(HaveSuffix("testutils/get_current_file_test.go"))
        Expect(f).To(HavePrefix("/"))
    })
})
