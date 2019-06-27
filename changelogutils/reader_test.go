package changelogutils_test

import (
	"context"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/changelogutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/versionutils"
	"github.com/solo-io/go-utils/vfsutils"
	"os"
	"time"
)
var _ = Describe("ReaderTest", func() {

	var (
		ctx = context.TODO()
		reader changelogutils.ChangelogReader
	)

	Context("happypath with github", func() {

		const (
			owner = "solo-io"
			repo = "testrepo"
			sha = "9065a9a84e286ea7f067f4fc240944b0a4d4c82a"
		)

		var (
			code vfsutils.MountedRepo
			entry = changelogutils.ChangelogEntry{
				Type: changelogutils.NEW_FEATURE,
				Description: "Now testrepo pushes rendered changelog to solo-docs on release builds.",
				IssueLink: "https://github.com/solo-io/testrepo/issues/9",
			}
			file = changelogutils.ChangelogFile{
				Entries: []*changelogutils.ChangelogEntry{&entry},
			}
		)

		BeforeEach(func() {
			client, err := githubutils.GetClient(ctx)
			Expect(err).NotTo(HaveOccurred())
			code = vfsutils.NewLazilyMountedRepo(client, owner, repo, sha)
			reader = changelogutils.NewChangelogReader(code)
		})

		It("can read changelog file", func() {
			changelogFile, err := reader.ReadChangelogFile(ctx, "changelog/v0.1.1/1.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(*changelogFile).To(BeEquivalentTo(file))
		})

		It("can get changelog", func() {
			changelog, err := reader.GetChangelogForTag(ctx, "v0.1.1")
			Expect(err).NotTo(HaveOccurred())
			expected := changelogutils.Changelog{
				Files: []*changelogutils.ChangelogFile{&file},
				Version: versionutils.NewVersion(0, 1, 1),
			}
			Expect(*changelog).To(BeEquivalentTo(expected))
		})
	})

	Context("edge cases with mocked mounted repo", func() {

		const (
			tag = "v0.0.1"
			changelogDir = "changelog/v0.0.1"
			summaryFile = "changelog/v0.0.1/summary.md"
			closingFile = "changelog/v0.0.1/closing.md"
			changelogFile = "changelog/v0.0.1/changelog.yaml"
		)

		var (
			ctrl *gomock.Controller
			mockCode *MockMountedRepo
			nestedErr = errors.Errorf("")
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(test)
			mockCode = NewMockMountedRepo(ctrl)
			reader = changelogutils.NewChangelogReader(mockCode)
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("errors on unexpected directory", func() {
			files := []os.FileInfo{getFileInfo("foo", true)}
			mockCode.EXPECT().
				ListFiles(ctx, changelogDir).
				Return(files, nil)

			expected := changelogutils.UnexpectedDirectoryError("foo", changelogDir)

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err.Error()).To(Equal(expected.Error()))
		})

		It("errors on reading summary", func() {
			files := []os.FileInfo{getFileInfo("summary.md", false)}
			mockCode.EXPECT().
				ListFiles(ctx, changelogDir).
				Return(files, nil)
			mockCode.EXPECT().
				GetFileContents(ctx, summaryFile).
				Return(nil, nestedErr)

			expected := changelogutils.UnableToReadSummaryFileError(nestedErr, summaryFile)

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err.Error()).To(Equal(expected.Error()))
		})

		It("errors on reading closing", func() {
			files := []os.FileInfo{getFileInfo("closing.md", false)}
			mockCode.EXPECT().
				ListFiles(ctx, changelogDir).
				Return(files, nil)
			mockCode.EXPECT().
				GetFileContents(ctx, closingFile).
				Return(nil, nestedErr)

			expected := changelogutils.UnableToReadClosingFileError(nestedErr, closingFile)

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err.Error()).To(Equal(expected.Error()))
		})

		It("errors on no entries in file", func() {
			files := []os.FileInfo{getFileInfo("changelog.yaml", false)}
			mockCode.EXPECT().
				ListFiles(ctx, changelogDir).
				Return(files, nil)
			mockCode.EXPECT().
				GetFileContents(ctx, changelogFile).
				Return([]byte(changelogNoEntries), nil)

			expected := changelogutils.NoEntriesInChangelogError(changelogFile)

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err.Error()).To(Equal(expected.Error()))
		})

		It("errors on parsing problem", func() {
			files := []os.FileInfo{getFileInfo("changelog.yaml", false)}
			mockCode.EXPECT().
				ListFiles(ctx, changelogDir).
				Return(files, nil)
			mockCode.EXPECT().
				GetFileContents(ctx, changelogFile).
				Return([]byte("invalid changelog"), nil)

			expected := changelogutils.UnableToParseChangelogError(nestedErr, changelogFile)

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err.Error()).To(ContainSubstring(expected.Error()))
		})

		It("errors on missing issue link", func() {
			files := []os.FileInfo{getFileInfo("changelog.yaml", false)}
			mockCode.EXPECT().
				ListFiles(ctx, changelogDir).
				Return(files, nil)
			mockCode.EXPECT().
				GetFileContents(ctx, changelogFile).
				Return([]byte(changelogMissingIssueLink), nil)

			expected := changelogutils.MissingIssueLinkError

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err).To(Equal(expected))
		})

		It("errors on missing description", func() {
			files := []os.FileInfo{getFileInfo("changelog.yaml", false)}
			mockCode.EXPECT().
				ListFiles(ctx, changelogDir).
				Return(files, nil)
			mockCode.EXPECT().
				GetFileContents(ctx, changelogFile).
				Return([]byte(changelogMissingDescription), nil)

			expected := changelogutils.MissingDescriptionError

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err).To(Equal(expected))
		})

		It("errors on missing owner", func() {
			files := []os.FileInfo{getFileInfo("changelog.yaml", false)}
			mockCode.EXPECT().
				ListFiles(ctx, changelogDir).
				Return(files, nil)
			mockCode.EXPECT().
				GetFileContents(ctx, changelogFile).
				Return([]byte(changelogMissingOwner), nil)

			expected := changelogutils.MissingOwnerError

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err).To(Equal(expected))
		})

		It("errors on missing repo", func() {
			files := []os.FileInfo{getFileInfo("changelog.yaml", false)}
			mockCode.EXPECT().
				ListFiles(ctx, changelogDir).
				Return(files, nil)
			mockCode.EXPECT().
				GetFileContents(ctx, changelogFile).
				Return([]byte(changelogMissingRepo), nil)

			expected := changelogutils.MissingRepoError

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err).To(Equal(expected))
		})

		It("errors on missing tag", func() {
			files := []os.FileInfo{getFileInfo("changelog.yaml", false)}
			mockCode.EXPECT().
				ListFiles(ctx, changelogDir).
				Return(files, nil)
			mockCode.EXPECT().
				GetFileContents(ctx, changelogFile).
				Return([]byte(changelogMissingTag), nil)

			expected := changelogutils.MissingTagError

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err).To(Equal(expected))
		})

	})




})

const (
	changelogNoEntries = `
changelog:
`
	changelogMissingIssueLink = `
changelog: 
  - type: FIX
    description: test
`
	changelogMissingDescription = `
changelog: 
  - type: FIX
    issueLink: test
`
	changelogMissingOwner = `
changelog: 
  - type: DEPENDENCY_BUMP
    dependencyRepo: foo
    dependencyTag: bar
`
	changelogMissingRepo = `
changelog: 
  - type: DEPENDENCY_BUMP
    dependencyOwner: foo
    dependencyTag: bar
`
	changelogMissingTag = `
changelog: 
  - type: DEPENDENCY_BUMP
    dependencyOwner: foo
    dependencyRepo: bar
`
)

func getFileInfo(name string, isDir bool) os.FileInfo {
	return &mockFileInfo{
		name: name,
		isDir: isDir,
	}
}

type mockFileInfo struct {
	name string
	isDir bool
}

func (i *mockFileInfo) Name() string {
	return i.name
}

func (i *mockFileInfo) Size() int64 {
	return 0
}

func (i *mockFileInfo) Mode() os.FileMode {
	return os.ModePerm
}

func (i *mockFileInfo) ModTime() time.Time {
	return time.Now()
}

func (i *mockFileInfo) IsDir() bool {
	return i.isDir
}

func (i *mockFileInfo) Sys() interface{} {
	return nil
}


