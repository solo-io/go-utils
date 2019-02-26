package changelogutils_test

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/solo-io/go-utils/changelogutils"
	"github.com/solo-io/go-utils/versionutils"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("ChangelogTest", func() {

	var _ = Context("GetProposedTag", func() {
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
			expectGetProposedTag("v0.0.2", tmpDir, "", "Versions v0.0.4 and v0.0.3 are both greater than latest tag v0.0.2")
			expectGetProposedTag("v0.0.4", tmpDir, "", "No version greater than v0.0.4 found")
			Expect(createSubdirs(changelogDir, "0.0.5")).To(BeNil())
			expectGetProposedTag("v0.0.5", tmpDir, "", "Directory name 0.0.5 is not valid, must be of the form 'vX.Y.Z'")
		})
	})

	It("can marshal changelog entries", func() {
		var clf changelogutils.ChangelogFile
		err := yaml.Unmarshal([]byte(mockChangelog), &clf)
		Expect(err).NotTo(HaveOccurred())
		_, err = yaml.Marshal(clf)
		Expect(err).NotTo(HaveOccurred())
	})

	var _ = Context("", func() {
		var fs afero.Fs
		createChangelogDir := func(tag string) {
			fs.MkdirAll(filepath.Join(changelogutils.ChangelogDirectory, tag), 0700)
		}
		writeChangelogFile := func(file *changelogutils.ChangelogFile, filename, tag string) {
			filepath := filepath.Join(changelogutils.ChangelogDirectory, tag, filename)
			bytes, err := yaml.Marshal(file)
			Expect(err).NotTo(HaveOccurred())
			afero.WriteFile(fs, filepath, bytes, 0700)
		}
		writeSummaryFile := func(summary, tag string) {
			filepath := filepath.Join(changelogutils.ChangelogDirectory, tag, changelogutils.SummaryFile)
			afero.WriteFile(fs, filepath, []byte(summary), 0700)
		}
		writeChangelog := func(changelog *changelogutils.Changelog) {
			tag := changelog.Version.String()
			createChangelogDir(tag)
			if changelog.Summary != "" {
				writeSummaryFile(changelog.Summary, tag)
			}
			for i, file := range changelog.Files {
				writeChangelogFile(file, fmt.Sprintf("%d.yaml", i), tag)
			}
		}
		getChangelog := func(tag, summary string, files ...*changelogutils.ChangelogFile) *changelogutils.Changelog{
			version, err := versionutils.ParseVersion(tag)
			Expect(err).NotTo(HaveOccurred())
			return &changelogutils.Changelog{
				Summary: summary,
				Version: version,
				Files: files,
			}
		}
		getChangelogFile := func(entries ...*changelogutils.ChangelogEntry) *changelogutils.ChangelogFile {
			return &changelogutils.ChangelogFile{
				Entries: entries,
			}
		}
		getStableApiChangelogFile := func(entries ...*changelogutils.ChangelogEntry) *changelogutils.ChangelogFile {
			return &changelogutils.ChangelogFile{
				Entries: entries,
				ReleaseStableApi: true,
			}
		}
		getEntry := func(entryType changelogutils.ChangelogEntryType, description, issue string) *changelogutils.ChangelogEntry {
			return &changelogutils.ChangelogEntry{
				Type: entryType,
				Description: description,
				IssueLink: issue,
			}
		}

		BeforeEach(func() {
			fs = afero.NewMemMapFs()
		})

		It("works", func() {
			tag := "v0.0.2"
			changelog := getChangelog(tag, "blah",
				getChangelogFile(
					getEntry(changelogutils.FIX, "fixes foo", "foo"),
					getEntry(changelogutils.FIX, "fixes bar", "bar"),
					getEntry(changelogutils.NEW_FEATURE, "adds baz", "baz")),
				getChangelogFile(getEntry(changelogutils.FIX, "fixes foo2", "foo2")),
				getChangelogFile(getEntry(changelogutils.NON_USER_FACING, "fixes foo3", "foo3")))
			writeChangelog(changelog)
			loadedChangelog, err := changelogutils.ComputeChangelog(fs, "v0.0.1", tag, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(loadedChangelog).To(BeEquivalentTo(changelog))
		})

		It("validates minor version should get bumped for breaking change", func() {
			tag := "v0.0.2"
			changelog := getChangelog(tag, "",
				getChangelogFile(getEntry(changelogutils.BREAKING_CHANGE, "fixes foo", "foo")))
			writeChangelog(changelog)
			_, err := changelogutils.ComputeChangelog(fs, "v0.0.1", tag, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("Expected version v0.1.0 to be next changelog version, found v0.0.2"))
		})

		It("validates major version should get bumped for breaking change", func() {
			tag := "v2.0.0"
			changelog := getChangelog(tag, "",
				getChangelogFile(getEntry(changelogutils.BREAKING_CHANGE, "fixes foo", "foo")))
			writeChangelog(changelog)
			loadedChangelog, err := changelogutils.ComputeChangelog(fs, "v1.2.2", tag, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(loadedChangelog).To(BeEquivalentTo(changelog))
		})

		It("validates no extra subdirectories are in the changelog directory", func() {
			tag := "v0.0.2"
			changelog := getChangelog(tag, "",
				getChangelogFile(getEntry(changelogutils.FIX, "fixes foo", "foo")))
			writeChangelog(changelog)
			fs.Mkdir(filepath.Join(changelogutils.ChangelogDirectory, tag, "foo"), 0700)
			_, err := changelogutils.ComputeChangelog(fs, "v0.0.1", tag, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("Unexpected directory foo in changelog directory changelog/v0.0.2"))
		})

		It("validates no extra files are in the changelog directory", func() {
			tag := "v0.0.2"
			changelog := getChangelog(tag, "",
				getChangelogFile(getEntry(changelogutils.FIX, "fixes foo", "foo")))
			writeChangelog(changelog)
			afero.WriteFile(fs, filepath.Join(changelogutils.ChangelogDirectory, tag, "foo"), []byte("invalid changelog"), 0700)
			_, err := changelogutils.ComputeChangelog(fs, "v0.0.1", tag, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("File changelog/v0.0.2/foo is not a valid changelog file"))
		})

		It("validates no extra files are in the changelog directory", func() {
			tag := "v0.0.2"
			changelog := getChangelog(tag, "",
				getChangelogFile(getEntry(changelogutils.FIX, "fixes foo", "foo")))
			writeChangelog(changelog)
			afero.WriteFile(fs, filepath.Join(changelogutils.ChangelogDirectory, tag, "foo"), []byte("invalid changelog"), 0700)
			_, err := changelogutils.ComputeChangelog(fs, "v0.0.1", tag, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("File changelog/v0.0.2/foo is not a valid changelog file"))
		})

		It("releasing stable API (v1.0.0) works", func() {
			tag := "v1.0.0"
			changelog := getChangelog(tag, "",
				getStableApiChangelogFile(getEntry(changelogutils.BREAKING_CHANGE, "fixes foo", "foo")))
			writeChangelog(changelog)
			loadedChangelog, err := changelogutils.ComputeChangelog(fs, "v0.0.1", tag, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(loadedChangelog).To(BeEquivalentTo(changelog))
		})

		It("releasing stable API must happen in v1.0.0 release", func() {
			tag := "v1.1.0"
			changelog := getChangelog(tag, "",
				getStableApiChangelogFile(getEntry(changelogutils.BREAKING_CHANGE, "fixes foo", "foo")))
			writeChangelog(changelog)
			_, err := changelogutils.ComputeChangelog(fs, "v1.0.1", tag, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("Changelog indicates this is a stable API release, which should be used only to indicate the release of v1.0.0, not v1.1.0"))
		})

		It("proposed version must be greater than latest", func() {
			tag := "v0.1.0"
			changelog := getChangelog(tag, "",
				getChangelogFile(getEntry(changelogutils.FIX, "fixes foo", "foo")))
			writeChangelog(changelog)
			_, err := changelogutils.ComputeChangelog(fs, "v0.2.0", tag, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("Proposed version v0.1.0 must be greater than latest version v0.2.0"))
		})

		It("checks that changelog entries have a description", func() {
			tag := "v0.3.0"
			changelog := getChangelog(tag, "",
				getChangelogFile(getEntry(changelogutils.FIX, "", "foo")))
			writeChangelog(changelog)
			_, err := changelogutils.ComputeChangelog(fs, "v0.2.0", tag, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("Changelog entries must have a description"))
		})

		It("checks that changelog entries have an issue link", func() {
			tag := "v0.3.0"
			changelog := getChangelog(tag, "",
				getChangelogFile(getEntry(changelogutils.FIX, "foo", "")))
			writeChangelog(changelog)
			_, err := changelogutils.ComputeChangelog(fs, "v0.2.0", tag, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("Changelog entries must have an issue link"))
		})
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
  issueLink: https://github.com/solo-io/testrepo/issues/9
- type: NEW_FEATURE
  description: "new feature"
  issue: https://github.com/solo-io/testrepo/issues/9
- type: BREAKING_CHANGE
  description: "It's a breaker"
  issueLink: https://github.com/solo-io/testrepo/issues/9
`