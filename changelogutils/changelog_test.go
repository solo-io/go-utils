package changelogutils_test

import (
	"github.com/ghodss/yaml"
	"github.com/solo-io/go-utils/changelogutils"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("ChangelogTest", func() {

	expectGetProposedTag := func(latestTag, changelogDir, tag, err string) {
		fs := afero.NewOsFs()
		actualTag, actualErr := changelogutils.GetProposedTag(fs, latestTag, changelogDir)
		Expect(actualTag).To(BeEquivalentTo(tag))
		if err == "" {
			Expect(actualErr).To(BeNil())
		} else {
			Expect(actualErr.Error()).To(BeEquivalentTo(err))
		}
	}

	It("works", func() {
		tmpDir := mustWriteTestDir()
		defer os.RemoveAll(tmpDir)
		changelogDir := filepath.Join(tmpDir, changelogutils.ChangelogDirectory)
		Expect(os.Mkdir(changelogDir, 0700)).To(BeNil())
		Expect(createSubdirs(changelogDir, "v0.0.1", "v0.0.2", "v0.0.3", "v0.0.4")).To(BeNil())
		expectGetProposedTag("v0.0.3", tmpDir, "v0.0.4", "")
	})

	It("can marshal changelog entries", func() {
		var clf changelogutils.ChangelogFile
		err := yaml.Unmarshal([]byte(mockChangelog), &clf)
		Expect(err).NotTo(HaveOccurred())
		_, err = yaml.Marshal(clf)
		Expect(err).NotTo(HaveOccurred())
	})
})

func createSubdirs(dir string, names ...string) error {
	for _, name := range names {
		subdir := filepath.Join(dir, name)
		err := os.Mkdir(subdir, 0700)
		if err != nil {
			return err
		}
	}
	return nil
}

func mustWriteTestDir() string {
	tmpDir, err := ioutil.TempDir("", "changelog-test-")
	Expect(err).NotTo(HaveOccurred())
	return tmpDir
}

var mockChangelog = `
changelog:
- type: FIX
  description: "fix 1"
- type: NEW_FEATURE
  description: "new feature"
- type: BREAKING_CHANGE
  description: "It's a breaker"
`