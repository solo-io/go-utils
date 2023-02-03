package versionutils_test

import (
    "io/ioutil"
    "os"
    "path/filepath"

    "github.com/solo-io/go-utils/versionutils/dep"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"

    . "github.com/solo-io/go-utils/versionutils"
)

var _ = Describe("Repo", func() {

    var (
        err         error
        tmpDir      string
        tomlFile    string
        tomlWrapper *TomlWrapper
        tomlContent = `
[[override]]
  name = "github.com/solo-io/gloo"
  branch = "master"

[[constraint]]
  name = "github.com/solo-io/service-mesh-hub"
  revision = "f1cdf253cb03da85ed2af456140c38eb90bcb53b"

[[constraint]]
  name = "github.com/solo-io/supergloo"
  version = "v0.3.25"
`
    )

    BeforeEach(func() {
        tmpDir, err = ioutil.TempDir("", "toml-parse-test-")
        Expect(err).NotTo(HaveOccurred())
        tomlFile = filepath.Join(tmpDir, "Gopkg.toml")
        err = ioutil.WriteFile(tomlFile, []byte(tomlContent), 0700)
        Expect(err).NotTo(HaveOccurred())

        tomlWrapper, err = ParseFullTomlFromDir(tmpDir)
        Expect(err).NotTo(HaveOccurred())
    })

    AfterEach(func() {
        _ = os.RemoveAll(tmpDir)
    })

    It("can get the version from when 'version' is specified", func() {
        version, err := GetDependencyVersionInfo("github.com/solo-io/supergloo", tomlWrapper)
        Expect(err).NotTo(HaveOccurred())
        Expect(version.Version).To(Equal("v0.3.25"))
        Expect(version.Type).To(Equal(dep.Version))
    })

    It("can get the version from when 'branch' is specified", func() {
        version, err := GetDependencyVersionInfo("github.com/solo-io/gloo", tomlWrapper)
        Expect(err).NotTo(HaveOccurred())
        Expect(version.Version).To(Equal("master"))
        Expect(version.Type).To(Equal(dep.Branch))
    })

    It("can get the version from when 'revision' is specified", func() {
        version, err := GetDependencyVersionInfo("github.com/solo-io/service-mesh-hub", tomlWrapper)
        Expect(err).NotTo(HaveOccurred())
        Expect(version.Version).To(Equal("f1cdf253cb03da85ed2af456140c38eb90bcb53b"))
        Expect(version.Type).To(Equal(dep.Revision))
    })

    Context("deprecated functions", func() {

        It("can get the version from when 'version' is specified", func() {
            version, err := GetTomlVersion("github.com/solo-io/supergloo", tomlWrapper)
            Expect(err).NotTo(HaveOccurred())
            Expect(version).To(Equal("v0.3.25"))
        })

        It("can get the version from when 'branch' is specified", func() {
            version, err := GetTomlVersion("github.com/solo-io/gloo", tomlWrapper)
            Expect(err).NotTo(HaveOccurred())
            Expect(version).To(Equal("master"))
        })

        It("can get the version from when 'revision' is specified", func() {
            version, err := GetTomlVersion("github.com/solo-io/service-mesh-hub", tomlWrapper)
            Expect(err).NotTo(HaveOccurred())
            Expect(version).To(Equal("f1cdf253cb03da85ed2af456140c38eb90bcb53b"))
        })
    })
})
