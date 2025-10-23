package changelogutils_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/changelogutils"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/versionutils"
	"github.com/spf13/afero"
)

var _ = Describe("ChangelogTest", func() {

	var _ = Context("GetProposedTag", func() {
		getProposedTag := func(latestTag, changelogDir, tag string) error {
			fs := afero.NewOsFs()
			actualTag, actualErr := changelogutils.GetProposedTag(fs, latestTag, changelogDir)
			Expect(actualTag).To(BeEquivalentTo(tag))
			return actualErr
		}

		It("works", func() {
			tmpDir := mustWriteTestDir()
			defer os.RemoveAll(tmpDir)
			changelogDir := filepath.Join(tmpDir, changelogutils.ChangelogDirectory)
			Expect(os.Mkdir(changelogDir, 0700)).To(BeNil())
			Expect(getProposedTag("v0.0.0", tmpDir, "v0.0.1")).To(BeNil())
			Expect(createSubdirs(changelogDir, "v0.0.1", "v0.0.2", "v0.0.3", "v0.0.4")).To(BeNil())
			Expect(getProposedTag("v0.0.3", tmpDir, "v0.0.4")).To(BeNil())
			Expect(changelogutils.IsMultipleVersionsFoundError(getProposedTag("v0.0.2", tmpDir, ""))).To(BeTrue())
			Expect(changelogutils.IsNoVersionFoundError(getProposedTag("v0.0.4", tmpDir, ""))).To(BeTrue())

			// test that we can switch between beta and rc releases
			Expect(createSubdirs(changelogDir, "v1.0.0-beta1", "v1.0.0-beta2")).To(BeNil())
			Expect(getProposedTag("v1.0.0-beta1", tmpDir, "v1.0.0-beta2")).To(BeNil())
			Expect(createSubdirs(changelogDir, "v1.0.0-rc1")).To(BeNil())
			Expect(getProposedTag("v1.0.0-beta2", tmpDir, "v1.0.0-rc1")).To(BeNil())
			Expect(createSubdirs(changelogDir, "v1.0.0-rc2")).To(BeNil())
			Expect(getProposedTag("v1.0.0-rc1", tmpDir, "")).NotTo(BeNil())

			// add a directory without 'v' prefix, which should be parsed as invalid
			Expect(createSubdirs(changelogDir, "1.0.0-beta3")).To(BeNil())
			Expect(changelogutils.IsInvalidDirectoryNameError(getProposedTag("v1.0.0-beta2", tmpDir, ""))).To(BeTrue())
		})
	})

	var _ = Context("Changelog marshaling", func() {

		It("can marshal changelog entries", func() {
			var clf changelogutils.ChangelogFile
			err := yaml.Unmarshal([]byte(mockChangelog), &clf)
			for _, value := range clf.Entries {
				Expect(value.Type.String()).NotTo(BeEmpty())
				Expect(value.Description).NotTo(BeEmpty())
				Expect(value.IssueLink).NotTo(BeEmpty())
				Expect(value.GetResolvesIssue()).To(BeTrue()) // default
			}
			Expect(err).NotTo(HaveOccurred())
			byt, err := yaml.Marshal(clf)
			Expect(err).NotTo(HaveOccurred())
			Expect(strings.Contains(string(byt), "releaseStableApi")).To(BeFalse())
		})

		It("can handle resolvesIssue set to false", func() {
			var clf changelogutils.ChangelogFile
			contents := `changelog:
- type: FIX
  description: foo
  issueLink: bar
  resolvesIssue: false`
			err := yaml.Unmarshal([]byte(contents), &clf)
			Expect(err).NotTo(HaveOccurred())
			boolValue := new(bool)
			*boolValue = false
			expected := changelogutils.ChangelogFile{
				Entries: []*changelogutils.ChangelogEntry{
					{
						Type:          changelogutils.FIX,
						Description:   "foo",
						IssueLink:     "bar",
						ResolvesIssue: boolValue,
					},
				},
			}
			Expect(clf).To(BeEquivalentTo(expected))
		})

		It("can handle resolvesIssue set to true", func() {
			var clf changelogutils.ChangelogFile
			contents := `changelog:
- type: FIX
  description: foo
  issueLink: bar
  resolvesIssue: true`
			err := yaml.Unmarshal([]byte(contents), &clf)
			Expect(err).NotTo(HaveOccurred())
			boolValue := new(bool)
			*boolValue = true
			expected := changelogutils.ChangelogFile{
				Entries: []*changelogutils.ChangelogEntry{
					{
						Type:          changelogutils.FIX,
						Description:   "foo",
						IssueLink:     "bar",
						ResolvesIssue: boolValue,
					},
				},
			}
			Expect(clf).To(BeEquivalentTo(expected))
		})
	})

	var _ = Context("Changelog computing and rendering", func() {
		var (
			fs afero.Fs

			boolean = true
			boolPtr = &boolean
		)
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
		writeClosingFile := func(closing, tag string) {
			filepath := filepath.Join(changelogutils.ChangelogDirectory, tag, changelogutils.ClosingFile)
			afero.WriteFile(fs, filepath, []byte(closing), 0700)
		}
		writeChangelog := func(changelog *changelogutils.Changelog) {
			tag := changelog.Version.String()
			createChangelogDir(tag)
			if changelog.Summary != "" {
				writeSummaryFile(changelog.Summary, tag)
			}
			if changelog.Closing != "" {
				writeClosingFile(changelog.Closing, tag)
			}
			for i, file := range changelog.Files {
				writeChangelogFile(file, fmt.Sprintf("%d.yaml", i), tag)
			}
		}
		getChangelog := func(tag, summary, closing string, files ...*changelogutils.ChangelogFile) *changelogutils.Changelog {
			version, err := versionutils.ParseVersion(tag)
			Expect(err).NotTo(HaveOccurred())
			return &changelogutils.Changelog{
				Summary: summary,
				Closing: closing,
				Version: version,
				Files:   files,
			}
		}
		getChangelogFile := func(entries ...*changelogutils.ChangelogEntry) *changelogutils.ChangelogFile {
			return &changelogutils.ChangelogFile{
				Entries: entries,
			}
		}
		getStableApiChangelogFile := func(entries ...*changelogutils.ChangelogEntry) *changelogutils.ChangelogFile {
			return &changelogutils.ChangelogFile{
				Entries:          entries,
				ReleaseStableApi: boolPtr,
			}
		}
		getEntry := func(entryType changelogutils.ChangelogEntryType, description, issue string) *changelogutils.ChangelogEntry {
			return &changelogutils.ChangelogEntry{
				Type:        entryType,
				Description: description,
				IssueLink:   issue,
			}
		}

		BeforeEach(func() {
			fs = afero.NewMemMapFs()
		})

		It("can compute changelog", func() {
			latestTag := "v0.0.1"
			newTag := "v0.0.2"
			changelog := getChangelog(newTag, "blah", "closing",
				getChangelogFile(
					getEntry(changelogutils.FIX, "fixes foo", "foo"),
					getEntry(changelogutils.FIX, "fixes bar", "bar"),
					getEntry(changelogutils.NEW_FEATURE, "adds baz", "baz")),
				getChangelogFile(getEntry(changelogutils.FIX, "fixes foo2", "foo2")),
				getChangelogFile(getEntry(changelogutils.NON_USER_FACING, "fixes foo3", "foo3")))
			writeChangelog(changelog)
			loadedChangelog, err := changelogutils.ComputeChangelogForNonRelease(fs, latestTag, newTag, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(loadedChangelog).To(BeEquivalentTo(changelog))
		})

		It("can compute changelog for first release", func() {
			latestTag := "v0.0.0"
			newTag := "v0.0.1"
			changelog := getChangelog(newTag, "blah", "closing",
				getChangelogFile(getEntry(changelogutils.FIX, "fixes foo", "foo")))
			writeChangelog(changelog)
			loadedChangelog, err := changelogutils.ComputeChangelogForNonRelease(fs, latestTag, newTag, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(loadedChangelog).To(BeEquivalentTo(changelog))
		})

		It("validates minor version should get bumped for breaking change", func() {
			tag := "v0.0.2"
			changelog := getChangelog(tag, "", "",
				getChangelogFile(getEntry(changelogutils.BREAKING_CHANGE, "fixes foo", "foo")))
			writeChangelog(changelog)
			_, err := changelogutils.ComputeChangelogForNonRelease(fs, "v0.0.1", tag, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("Expected version v0.1.0 to be next changelog version, found v0.0.2"))
		})

		It("validates major version should get bumped for breaking change", func() {
			tag := "v2.0.0"
			changelog := getChangelog(tag, "", "",
				getChangelogFile(getEntry(changelogutils.BREAKING_CHANGE, "fixes foo", "foo")))
			writeChangelog(changelog)
			loadedChangelog, err := changelogutils.ComputeChangelogForNonRelease(fs, "v1.2.2", tag, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(loadedChangelog).To(BeEquivalentTo(changelog))
		})

		It("validates no extra subdirectories are in the changelog directory", func() {
			tag := "v0.0.2"
			changelog := getChangelog(tag, "", "",
				getChangelogFile(getEntry(changelogutils.FIX, "fixes foo", "foo")))
			writeChangelog(changelog)
			fs.Mkdir(filepath.Join(changelogutils.ChangelogDirectory, tag, "foo"), 0700)
			_, err := changelogutils.ComputeChangelogForNonRelease(fs, "v0.0.1", tag, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("Unexpected directory foo in changelog directory changelog/v0.0.2"))
		})

		It("validates no extra files are in the changelog directory", func() {
			tag := "v0.0.2"
			changelog := getChangelog(tag, "", "",
				getChangelogFile(getEntry(changelogutils.FIX, "fixes foo", "foo")))
			writeChangelog(changelog)
			afero.WriteFile(fs, filepath.Join(changelogutils.ChangelogDirectory, tag, "foo"), []byte("invalid changelog"), 0700)
			_, err := changelogutils.ComputeChangelogForNonRelease(fs, "v0.0.1", tag, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("File v0.0.2/foo is not a valid changelog file. Error: error unmarshaling JSON: json: cannot unmarshal string into Go value of type changelogutils.ChangelogFile"))
		})

		It("validates no extra files are in the changelog directory", func() {
			tag := "v0.0.2"
			changelog := getChangelog(tag, "", "",
				getChangelogFile(getEntry(changelogutils.FIX, "fixes foo", "foo")))
			writeChangelog(changelog)
			afero.WriteFile(fs, filepath.Join(changelogutils.ChangelogDirectory, tag, "foo"), []byte("invalid changelog"), 0700)
			_, err := changelogutils.ComputeChangelogForNonRelease(fs, "v0.0.1", tag, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("File v0.0.2/foo is not a valid changelog file. Error: error unmarshaling JSON: json: cannot unmarshal string into Go value of type changelogutils.ChangelogFile"))
		})

		It("releasing rc (v1.0.0-rc1) works", func() {
			tag := "v1.0.0-rc1"
			changelog := getChangelog(tag, "", "",
				getChangelogFile(getEntry(changelogutils.BREAKING_CHANGE, "fixes foo", "foo")))
			writeChangelog(changelog)
			loadedChangelog, err := changelogutils.ComputeChangelogForNonRelease(fs, "v0.0.1", tag, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(loadedChangelog).To(BeEquivalentTo(changelog))
		})

		It("releasing v1.0.0 after v1.0.0-rc1 works", func() {
			tag := "v1.0.0"
			changelog := getChangelog(tag, "", "",
				getStableApiChangelogFile(getEntry(changelogutils.BREAKING_CHANGE, "fixes foo", "foo")))
			writeChangelog(changelog)
			loadedChangelog, err := changelogutils.ComputeChangelogForNonRelease(fs, "v1.0.0-rc1", tag, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(loadedChangelog).To(BeEquivalentTo(changelog))
		})

		It("releasing stable API (v1.0.0) works", func() {
			tag := "v1.0.0"
			changelog := getChangelog(tag, "", "",
				getStableApiChangelogFile(getEntry(changelogutils.BREAKING_CHANGE, "fixes foo", "foo")))
			writeChangelog(changelog)
			loadedChangelog, err := changelogutils.ComputeChangelogForNonRelease(fs, "v0.0.1", tag, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(loadedChangelog).To(BeEquivalentTo(changelog))
		})

		// tests a deprecated version of the function that enforces stable releases in v1.0.0
		// the new changelog validator only enforces that stable APIs are released in versions >= v1.0.0
		It("releasing stable API must happen in v1.0.0 release", func() {
			tag := "v1.1.0"
			changelog := getChangelog(tag, "", "",
				getStableApiChangelogFile(getEntry(changelogutils.BREAKING_CHANGE, "fixes foo", "foo")))
			writeChangelog(changelog)
			_, err := changelogutils.ComputeChangelogForNonRelease(fs, "v1.0.1", tag, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("Changelog indicates this is a stable API release, which should be used only to indicate the release of v1.0.0, not v1.1.0"))
		})

		It("proposed version must be greater than latest", func() {
			tag := "v0.1.0"
			changelog := getChangelog(tag, "", "",
				getChangelogFile(getEntry(changelogutils.FIX, "fixes foo", "foo")))
			writeChangelog(changelog)
			_, err := changelogutils.ComputeChangelogForNonRelease(fs, "v0.2.0", tag, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("Proposed version v0.1.0 must be greater than latest version v0.2.0"))
		})

		It("checks that changelog entries have a description", func() {
			tag := "v0.3.0"
			changelog := getChangelog(tag, "", "",
				getChangelogFile(getEntry(changelogutils.FIX, "", "foo")))
			writeChangelog(changelog)
			_, err := changelogutils.ComputeChangelogForNonRelease(fs, "v0.2.0", tag, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("Changelog entries must have a description"))
		})

		It("checks that changelog entries have an issue link", func() {
			tag := "v0.3.0"
			changelog := getChangelog(tag, "", "",
				getChangelogFile(getEntry(changelogutils.FIX, "foo", "")))
			writeChangelog(changelog)
			_, err := changelogutils.ComputeChangelogForNonRelease(fs, "v0.2.0", tag, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("Changelog entries must have an issue link"))
		})

		It("can render changelog", func() {
			changelog := getChangelog("v0.0.1", "blah", "closing",
				getChangelogFile(
					getEntry(changelogutils.FIX, "fixes foo    ", "  foo  "), // testing trim space
					getEntry(changelogutils.BREAKING_CHANGE, "fixes bar", "bar"),
					getEntry(changelogutils.NEW_FEATURE, "adds baz", "baz")),
				getChangelogFile(getEntry(changelogutils.FIX, "fixes foo2", "foo2")),
				getChangelogFile(getEntry(changelogutils.NON_USER_FACING, "fixes foo3", "foo3")))
			output := changelogutils.GenerateChangelogMarkdown(changelog)
			expected := `blah

**Breaking Changes**

- fixes bar (bar)

**New Features**

- adds baz (baz)

**Fixes**

- fixes foo (foo)
- fixes foo2 (foo2)

closing

`
			Expect(output).To(BeEquivalentTo(expected))
		})

		It("can render changelog with only fixes and closing", func() {
			changelog := getChangelog("v0.0.1", "", "closing",
				getChangelogFile(getEntry(changelogutils.FIX, "fixes foo2", "foo2")))
			output := changelogutils.GenerateChangelogMarkdown(changelog)
			expected := `**Fixes**

- fixes foo2 (foo2)

closing

`
			Expect(output).To(BeEquivalentTo(expected))
		})

		It("can render changelog with only summary", func() {
			changelog := getChangelog("v0.0.1", "blah", "",
				getChangelogFile(getEntry(changelogutils.NON_USER_FACING, "fixes foo2", "foo2")))
			output := changelogutils.GenerateChangelogMarkdown(changelog)
			expected := "blah\n\n"
			Expect(output).To(BeEquivalentTo(expected))
		})

		It("can render changelog with no user-facing content", func() {
			changelog := getChangelog("v0.0.1", "", "",
				getChangelogFile(getEntry(changelogutils.NON_USER_FACING, "fixes foo2", "foo2")))
			output := changelogutils.GenerateChangelogMarkdown(changelog)
			expected := "This release contained no user-facing changes.\n\n"
			Expect(output).To(BeEquivalentTo(expected))
		})

		It("allows non user facing changes to not have a description or link", func() {
			changelog := getChangelog("v0.0.1", "", "",
				getChangelogFile(getEntry(changelogutils.NON_USER_FACING, "", "")))
			output := changelogutils.GenerateChangelogMarkdown(changelog)
			expected := "This release contained no user-facing changes.\n\n"
			Expect(output).To(BeEquivalentTo(expected))
		})
	})

	Context("Check for presence of changelog", func() {
		It("passes on repo with changelog", func() {
			ctx := context.Background()
			client, err := githubutils.GetClient(ctx)
			Expect(err).NotTo(HaveOccurred())
			hasChangelog, err := changelogutils.RefHasChangelog(ctx, client, "solo-io", "testrepo", "master")
			Expect(err).NotTo(HaveOccurred())
			Expect(hasChangelog).To(BeTrue())
		})

		It("fails on repo with no changelog", func() {
			ctx := context.Background()
			client, err := githubutils.GetClient(ctx)
			Expect(err).NotTo(HaveOccurred())
			hasChangelog, err := changelogutils.RefHasChangelog(ctx, client, "solo-io", "solo-docs", "master")
			Expect(err).NotTo(HaveOccurred())
			Expect(hasChangelog).To(BeFalse())
		})
	})

	Context("Summary documentation generation", func() {
		It("produces the expected output", func() {
			ctx := context.Background()
			repoRootPath := ".."
			owner := "solo-io"
			repo := "go-utils"
			changelogDirPath := "changelog"

			w := bytes.NewBuffer([]byte{})
			err := changelogutils.GenerateChangelogFromLocalDirectory(ctx, repoRootPath, owner, repo, changelogDirPath, w)
			Expect(err).NotTo(HaveOccurred())
			// testing against a substring since the full value will change with each new changelog
			// this substring should never change unless we change our changelog formatting
			// it covers the sorting concern, showing that 0.2.12 is indeed greater than 0.2.8
			Expect(w.String()).To(ContainSubstring(`
### v0.2.12

**Fixes**

- No longer create unwanted nested directory when pushing solo-kit docs for the first time. (https://github.com/solo-io/go-utils/issues/43)


### v0.2.11

**Fixes**

- Fixes the sha upload in ` + "`" + `UploadReleaseAssetsCli` + "`" + ` to upload a checksum for ` + "`" + `foo` + "`" + ` that matches the output of ` + "`" + `shasum -a 256 foo &gt; foo.sha256` + "`" + `. (https://github.com/solo-io/go-utils/issues/41)


### v0.2.10

**New Features**

- A utility CLI has been added for uploading release artifacts to github, to replace the old shell script. See the [readme](https://github.com/solo-io/go-utils/tree/master/githubutils) for more information. (https://github.com/solo-io/go-utils/issues/38)

**Fixes**

- PushDocsCli no longer errors on the initial push when the destination directory doesn&#39;t exist. (https://github.com/solo-io/go-utils/issues/40)


### v0.2.9

**New Features**

- The docs push utility now supports automated CLI docs. (https://github.com/solo-io/go-utils/issues/33)
- The docs can support API or CLI docs that are not in the root of the repo. (https://github.com/solo-io/go-utils/issues/33)
- The docs CLI library now includes the full CLI, so projects can execute docs push in 1 line. (https://github.com/solo-io/go-utils/issues/33)
- Moves common documentation generation to a shared lib (https://github.com/solo-io/go-utils/issues/35)


### v0.2.8

**New Features**

- Changelog now enabled for this repo. (https://github.com/solo-io/go-utils/issues/31)

**Fixes**

- Markdown generation now always ends in two new lines. (https://github.com/solo-io/go-utils/issues/30)
`))

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
	tmpDir, err := os.MkdirTemp("", "changelog-test-")
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
  issueLink: https://github.com/solo-io/testrepo/issues/9
- type: BREAKING_CHANGE
  description: "It's a breaker"
  issueLink: https://github.com/solo-io/testrepo/issues/9
`
