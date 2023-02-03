package vfsutils_test

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "github.com/solo-io/go-utils/vfsutils"
    "github.com/spf13/afero"
)

var _ = Describe("vfsutils", func() {
    Context("MountTar", func() {
        It("works", func() {
            fs := afero.NewMemMapFs()
            dir, err := vfsutils.MountTar(fs, "https://github.com/XiaoMi/naftis/releases/download/0.1.4-rc6/manifest.tar.gz")
            Expect(err).NotTo(HaveOccurred())
            Expect(dir).NotTo(BeEquivalentTo(""))
            files, err := afero.ReadDir(fs, dir)
            Expect(err).NotTo(HaveOccurred())
            Expect(len(files)).To(BeEquivalentTo(3))
        })
    })
})
