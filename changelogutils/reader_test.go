package changelogutils_test

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/changelogutils"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/versionutils"
	"github.com/solo-io/go-utils/vfsutils"
)

var _ = Describe("ReaderTest", func() {

	var (
		ctx    = context.Background()
		reader changelogutils.ChangelogReader
	)

	Context("happypath with github", func() {

		const (
			owner = "solo-io"
			repo  = "testrepo"
			sha   = "9065a9a84e286ea7f067f4fc240944b0a4d4c82a"
		)

		var (
			code  vfsutils.MountedRepo
			entry = changelogutils.ChangelogEntry{
				Type:        changelogutils.NEW_FEATURE,
				Description: "Now testrepo pushes rendered changelog to solo-docs on release builds.",
				IssueLink:   "https://github.com/solo-io/testrepo/issues/9",
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
				Files:   []*changelogutils.ChangelogFile{&file},
				Version: versionutils.NewVersion(0, 1, 1, "", 0),
			}
			Expect(*changelog).To(BeEquivalentTo(expected))
		})
	})

	Context("edge cases with mocked mounted repo", func() {

		const (
			tag          = "v0.0.1"
			changelogDir = "changelog/v0.0.1"
		)

		var (
			ctrl      *gomock.Controller
			mockCode  *MockMountedRepo
			nestedErr = eris.Errorf("")
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(test)
			mockCode = NewMockMountedRepo(ctrl)
			reader = changelogutils.NewChangelogReader(mockCode)
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("errors on listing directory", func() {
			mockCode.EXPECT().
				ListFiles(ctx, changelogDir).
				Return(nil, nestedErr)
			mockCode.EXPECT().
				GetFileContents(ctx, "changelog/validation.yaml").
				Return([]byte(""), nil)

			expected := changelogutils.UnableToListFilesError(nestedErr, changelogDir)
			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err.Error()).To(Equal(expected.Error()))
		})

		It("errors on unexpected directory", func() {
			files := []os.FileInfo{getFileInfo("foo", true)}
			mockCode.EXPECT().
				ListFiles(ctx, changelogDir).
				Return(files, nil)
			mockCode.EXPECT().
				GetFileContents(ctx, "changelog/validation.yaml").
				Return([]byte(""), nil)

			expected := changelogutils.UnexpectedDirectoryError("foo", changelogDir)

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err.Error()).To(Equal(expected.Error()))
		})

		setupMockChangelogDir := func(filename, contents string, err error) string {
			path := filepath.Join(changelogDir, filename)
			files := []os.FileInfo{getFileInfo(filename, false)}
			mockCode.EXPECT().
				ListFiles(ctx, changelogDir).
				Return(files, nil)
			mockCode.EXPECT().
				GetFileContents(ctx, path).
				Return([]byte(contents), err)
			mockCode.EXPECT().
				GetFileContents(ctx, "changelog/validation.yaml").
				Return([]byte(""), nil)
			return path
		}

		It("errors on reading summary", func() {
			path := setupMockChangelogDir(changelogutils.SummaryFile, "", nestedErr)
			expected := changelogutils.UnableToReadSummaryFileError(nestedErr, path)

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err.Error()).To(Equal(expected.Error()))
		})

		It("errors on reading closing", func() {
			path := setupMockChangelogDir(changelogutils.ClosingFile, "", nestedErr)
			expected := changelogutils.UnableToReadClosingFileError(nestedErr, path)

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err.Error()).To(Equal(expected.Error()))
		})

		It("errors on no entries in file", func() {
			path := setupMockChangelogDir("changelog.yaml", changelogNoEntries, nil)
			expected := changelogutils.NoEntriesInChangelogError(path)

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err.Error()).To(Equal(expected.Error()))
		})

		It("errors on parsing problem", func() {
			path := setupMockChangelogDir("changelog.yaml", "invalid changelog", nil)
			expected := changelogutils.UnableToParseChangelogError(nestedErr, path)

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err.Error()).To(ContainSubstring(expected.Error()))
		})

		It("errors on missing issue link", func() {
			setupMockChangelogDir("changelog.yaml", changelogMissingIssueLink, nil)
			expected := changelogutils.MissingIssueLinkError

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err).To(Equal(expected))
		})

		It("errors on missing description", func() {
			setupMockChangelogDir("changelog.yaml", changelogMissingDescription, nil)
			expected := changelogutils.MissingDescriptionError

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err).To(Equal(expected))
		})

		It("errors on missing owner", func() {
			setupMockChangelogDir("changelog.yaml", changelogMissingOwner, nil)
			expected := changelogutils.MissingOwnerError

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err).To(Equal(expected))
		})

		It("errors on missing repo", func() {
			setupMockChangelogDir("changelog.yaml", changelogMissingRepo, nil)
			expected := changelogutils.MissingRepoError

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err).To(Equal(expected))
		})

		It("errors on missing tag", func() {
			setupMockChangelogDir("changelog.yaml", changelogMissingTag, nil)
			expected := changelogutils.MissingTagError

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(changelog).To(BeNil())
			Expect(err).To(Equal(expected))
		})

		It("can get complex changelog", func() {
			files := []os.FileInfo{
				getFileInfo("summary.md", false),
				getFileInfo("closing.md", false),
				getFileInfo("1.yaml", false),
				getFileInfo("2.yaml", false),
				getFileInfo("3.yaml", false),
				getFileInfo("4.yaml", false),
				getFileInfo("5.yaml", false),
			}
			mockCode.EXPECT().
				ListFiles(ctx, changelogDir).
				Return(files, nil)
			mockCode.EXPECT().
				GetFileContents(ctx, "changelog/validation.yaml").
				Return([]byte(""), nil)
			mockCode.EXPECT().
				GetFileContents(ctx, filepath.Join(changelogDir, "1.yaml")).
				Return([]byte(validChangelog1), nil)
			mockCode.EXPECT().
				GetFileContents(ctx, filepath.Join(changelogDir, "2.yaml")).
				Return([]byte(validChangelog2), nil)
			mockCode.EXPECT().
				GetFileContents(ctx, filepath.Join(changelogDir, "3.yaml")).
				Return([]byte(validChangelog3), nil)
			mockCode.EXPECT().
				GetFileContents(ctx, filepath.Join(changelogDir, "4.yaml")).
				Return([]byte(validUpgradeChangelog), nil)
			mockCode.EXPECT().
				GetFileContents(ctx, filepath.Join(changelogDir, "5.yaml")).
				Return([]byte(validHelmChangelog), nil)
			mockCode.EXPECT().
				GetFileContents(ctx, filepath.Join(changelogDir, "summary.md")).
				Return([]byte("summary"), nil)
			mockCode.EXPECT().
				GetFileContents(ctx, filepath.Join(changelogDir, "closing.md")).
				Return([]byte("closing"), nil)

			resolvesIssueFalse := false
			expected := changelogutils.Changelog{
				Files: []*changelogutils.ChangelogFile{
					{
						Entries: []*changelogutils.ChangelogEntry{
							{
								Type:        changelogutils.FIX,
								Description: "foo1",
								IssueLink:   "bar1",
							},
							{
								Type:        changelogutils.NEW_FEATURE,
								Description: "foo2",
								IssueLink:   "bar2",
							},
						},
					},
					{
						Entries: []*changelogutils.ChangelogEntry{
							{
								Type:        changelogutils.NON_USER_FACING,
								Description: "foo3",
							},
							{
								Type:          changelogutils.FIX,
								Description:   "foo4",
								IssueLink:     "bar4",
								ResolvesIssue: &resolvesIssueFalse,
							},
						},
					},
					{
						Entries: []*changelogutils.ChangelogEntry{
							{
								Type:            changelogutils.DEPENDENCY_BUMP,
								DependencyOwner: "foo",
								DependencyRepo:  "bar",
								DependencyTag:   "baz",
							},
						},
					},
					{
						Entries: []*changelogutils.ChangelogEntry{
							{
								Type:        changelogutils.UPGRADE,
								Description: "foo5",
								IssueLink:   "bar5",
							},
						},
					},
					{
						Entries: []*changelogutils.ChangelogEntry{
							{
								Type:        changelogutils.HELM,
								Description: "foo6",
								IssueLink:   "bar6",
							},
						},
					},
				},
				Version: versionutils.NewVersion(0, 0, 1, "", 0),
				Summary: "summary",
				Closing: "closing",
			}

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(err).To(BeNil())
			Expect(*changelog).To(BeEquivalentTo(expected))
		})

		It("can handle v1.0.0-rc1", func() {
			rcTag := "v1.0.0-rc1"
			rcDir := "changelog/" + rcTag
			files := []os.FileInfo{
				getFileInfo("1.yaml", false),
			}
			mockCode.EXPECT().
				ListFiles(ctx, rcDir).
				Return(files, nil)
			mockCode.EXPECT().
				GetFileContents(ctx, "changelog/validation.yaml").
				Return([]byte(""), nil)
			mockCode.EXPECT().
				GetFileContents(ctx, filepath.Join(rcDir, "1.yaml")).
				Return([]byte(validChangelog3), nil)

			expected := changelogutils.Changelog{
				Files: []*changelogutils.ChangelogFile{
					{
						Entries: []*changelogutils.ChangelogEntry{
							{
								Type:            changelogutils.DEPENDENCY_BUMP,
								DependencyOwner: "foo",
								DependencyRepo:  "bar",
								DependencyTag:   "baz",
							},
						},
					},
				},
				Version: versionutils.NewVersion(1, 0, 0, "rc", 1),
			}

			changelog, err := reader.GetChangelogForTag(ctx, "v1.0.0-rc1")
			Expect(err).To(BeNil())
			Expect(*changelog).To(BeEquivalentTo(expected))
		})

		It("can handle release stable api", func() {
			files := []os.FileInfo{
				getFileInfo("1.yaml", false),
			}
			mockCode.EXPECT().
				ListFiles(ctx, changelogDir).
				Return(files, nil)
			mockCode.EXPECT().
				GetFileContents(ctx, "changelog/validation.yaml").
				Return([]byte(""), nil)
			mockCode.EXPECT().
				GetFileContents(ctx, filepath.Join(changelogDir, "1.yaml")).
				Return([]byte(validStableReleaseChangelog), nil)

			releaseStableApiTrue := true
			expected := changelogutils.Changelog{
				Files: []*changelogutils.ChangelogFile{
					{
						Entries: []*changelogutils.ChangelogEntry{
							{
								Type: changelogutils.NON_USER_FACING,
							},
						},
						ReleaseStableApi: &releaseStableApiTrue,
					},
				},
				Version: versionutils.NewVersion(0, 0, 1, "", 0),
			}

			changelog, err := reader.GetChangelogForTag(ctx, tag)
			Expect(err).To(BeNil())
			Expect(*changelog).To(BeEquivalentTo(expected))
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

	validChangelog1 = `
changelog:
  - type: FIX
    description: foo1
    issueLink: bar1
  - type: NEW_FEATURE
    description: foo2
    issueLink: bar2
`

	validChangelog2 = `
changelog:
  - type: NON_USER_FACING
    description: foo3
  - type: FIX
    description: foo4
    issueLink: bar4
    resolvesIssue: false
`

	validChangelog3 = `
changelog:
  - type: DEPENDENCY_BUMP
    dependencyOwner: foo
    dependencyRepo: bar
    dependencyTag: baz
`

	validBreakingChangelog = `
changelog:
  - type: BREAKING_CHANGE
    description: foo
    issueLink: bar
`

	validNewFeatureChangelog = `
changelog:
  - type: NEW_FEATURE
    description: cool new feature
    issueLink: http://issue
`

	validNonBreakingNorNewFeatureChangelog = `
changelog:
  - type: NON_USER_FACING
  - type: DEPENDENCY_BUMP
    dependencyOwner: foo
    dependencyRepo: bar
    dependencyTag: baz
  - type: UPGRADE
    description: foo5
    issueLink: bar5
  - type: HELM
    description: foo6
    issueLink: bar6
  - type: FIX
    description: foo1
    issueLink: bar1
`

	validStableReleaseChangelog = `
changelog:
  - type: NON_USER_FACING
releaseStableApi: true
`

	validUpgradeChangelog = `
changelog:
  - type: UPGRADE
    description: foo5
    issueLink: bar5
`

	validHelmChangelog = `
changelog:
  - type: HELM
    description: foo6
    issueLink: bar6
`
)

func getFileInfo(name string, isDir bool) os.FileInfo {
	return &mockFileInfo{
		name:  name,
		isDir: isDir,
	}
}

type mockFileInfo struct {
	name  string
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
