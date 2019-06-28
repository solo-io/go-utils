package changelogutils_test

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/google/go-github/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/changelogutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/githubutils"
	"os"
	"path/filepath"
)

var _ = Describe("github utils", func() {

	const (
		base = "base"
		sha = "sha"
	)

	var (
		ctrl *gomock.Controller
		repoClient *MockRepoClient
		code       *MockMountedRepo
		validator changelogutils.ChangelogValidator
		ctx = context.Background()
		nestedErr = errors.Errorf("")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(test)
		code = NewMockMountedRepo(ctrl)
		repoClient = NewMockRepoClient(ctrl)
		validator = changelogutils.NewChangelogValidator(repoClient, code, base)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("should check changelog", func() {
		It("should check if master has changelog", func() {
			repoClient.EXPECT().
				DirectoryExists(ctx, changelogutils.MasterBranch, changelogutils.ChangelogDirectory).
				Return(true, nil)

			check, err := validator.ShouldCheckChangelog(ctx)
			Expect(err).To(BeNil())
			Expect(check).To(BeTrue())
		})

		It("should not check if error looking at master", func() {
			repoClient.EXPECT().
				DirectoryExists(ctx, changelogutils.MasterBranch, changelogutils.ChangelogDirectory).
				Return(false, nestedErr)

			check, err := validator.ShouldCheckChangelog(ctx)
			Expect(err).To(Equal(nestedErr))
			Expect(check).To(BeFalse())
		})

		It("should check if sha has changelog but master doesn't", func() {
			repoClient.EXPECT().
				DirectoryExists(ctx, changelogutils.MasterBranch, changelogutils.ChangelogDirectory).
				Return(false, nil)
			code.EXPECT().GetSha().Return(sha)
			repoClient.EXPECT().
				DirectoryExists(ctx, sha, changelogutils.ChangelogDirectory).
				Return(true, nil)

			check, err := validator.ShouldCheckChangelog(ctx)
			Expect(err).To(BeNil())
			Expect(check).To(BeTrue())
		})

		It("shouldn't check if sha and master don't have changelog", func() {
			repoClient.EXPECT().
				DirectoryExists(ctx, changelogutils.MasterBranch, changelogutils.ChangelogDirectory).
				Return(false, nil)
			code.EXPECT().GetSha().Return(sha)
			repoClient.EXPECT().
				DirectoryExists(ctx, sha, changelogutils.ChangelogDirectory).
				Return(false, nil)

			check, err := validator.ShouldCheckChangelog(ctx)
			Expect(err).To(BeNil())
			Expect(check).To(BeFalse())
		})

		It("shouldn't check if error looking at sha", func() {
			repoClient.EXPECT().
				DirectoryExists(ctx, changelogutils.MasterBranch, changelogutils.ChangelogDirectory).
				Return(false, nil)
			code.EXPECT().GetSha().Return(sha)
			repoClient.EXPECT().
				DirectoryExists(ctx, sha, changelogutils.ChangelogDirectory).
				Return(false, nestedErr)

			check, err := validator.ShouldCheckChangelog(ctx)
			Expect(err).To(Equal(nestedErr))
			Expect(check).To(BeFalse())
		})
	})

	Context("validate changelog", func() {

		const (
			filename1 = "1.yaml"
			filename2 = "2.yaml"
			filename3 = "3.yaml"
			tag = "v0.5.1"
			nextTag = "v0.5.2"
		)

		var (
			path1 = filepath.Join(changelogutils.ChangelogDirectory, tag, filename1)
			path2 = filepath.Join(changelogutils.ChangelogDirectory, tag, filename2)
			added = githubutils.COMMIT_FILE_STATUS_ADDED
			unexpectedFile = mockFileInfo{isDir: false, name: "unexpected"}

			path3 = filepath.Join(changelogutils.ChangelogDirectory, nextTag, filename3)
		)

		getChangelogDir := func(tag string) os.FileInfo {
			return &mockFileInfo{name: tag, isDir: true}
		}

		BeforeEach(func() {
			// so should check returns true
			repoClient.EXPECT().
				DirectoryExists(ctx, changelogutils.MasterBranch, changelogutils.ChangelogDirectory).
				Return(true, nil)
			code.EXPECT().GetSha().Return(sha)
		})

		It("propagates error comparing commits", func() {
			repoClient.EXPECT().
				CompareCommits(ctx, base, sha).
				Return(nil, nestedErr)
			file, err := validator.ValidateChangelog(ctx)
			Expect(file).To(BeNil())
			Expect(err).To(Equal(nestedErr))
		})

		It("errors when no changelog file added", func() {
			cc := github.CommitsComparison{}
			repoClient.EXPECT().
				CompareCommits(ctx, base, sha).
				Return(&cc, nil)

			expected := changelogutils.NoChangelogFileAddedError
			file, err := validator.ValidateChangelog(ctx)
			Expect(file).To(BeNil())
			Expect(err).To(Equal(expected))
		})

		It("errors when more than one changelog file added", func() {
			file1 := github.CommitFile{Filename: &path1, Status: &added}
			file2 := github.CommitFile{Filename: &path2, Status: &added}
			cc := github.CommitsComparison{Files: []github.CommitFile{file1, file2}}
			repoClient.EXPECT().
				CompareCommits(ctx, base, sha).
				Return(&cc, nil)

			expected := changelogutils.TooManyChangelogFilesAddedError(2)
			file, err := validator.ValidateChangelog(ctx)
			Expect(file).To(BeNil())
			Expect(err.Error()).To(Equal(expected.Error()))
		})

		It("errors when getting changelog file contents fails", func() {
			file1 := github.CommitFile{Filename: &path1, Status: &added}
			cc := github.CommitsComparison{Files: []github.CommitFile{file1}}
			repoClient.EXPECT().
				CompareCommits(ctx, base, sha).
				Return(&cc, nil)
			code.EXPECT().
				GetFileContents(ctx, path1).
				Return(nil, nestedErr)

			file, err := validator.ValidateChangelog(ctx)
			Expect(file).To(BeNil())
			Expect(err).To(Equal(nestedErr))
		})

		Context("validating proposed tag", func() {
			BeforeEach(func() {
				file1 := github.CommitFile{Filename: &path1, Status: &added}
				cc := github.CommitsComparison{Files: []github.CommitFile{file1}}
				repoClient.EXPECT().
					CompareCommits(ctx, base, sha).
					Return(&cc, nil)
				code.EXPECT().
					GetFileContents(ctx, path1).
					Return([]byte(validChangelog1), nil)
			})

			It("propagates error listing releases", func() {
				repoClient.EXPECT().
					FindLatestReleaseTagIncudingPrerelease(ctx).
					Return("", nestedErr)

				expected := changelogutils.ListReleasesError(nestedErr)
				file, err := validator.ValidateChangelog(ctx)
				Expect(file).To(BeNil())
				Expect(err.Error()).To(Equal(expected.Error()))
			})

			It("errors when unexpected file in changelog directory", func() {
				repoClient.EXPECT().
					FindLatestReleaseTagIncudingPrerelease(ctx).
					Return(tag, nil)
				code.EXPECT().
					ListFiles(ctx, changelogutils.ChangelogDirectory).
					Return([]os.FileInfo{&unexpectedFile}, nil)

				expected := changelogutils.UnexpectedFileInChangelogDirectoryError(unexpectedFile.name)
				file, err := validator.ValidateChangelog(ctx)
				Expect(file).To(BeNil())
				Expect(err.Error()).To(Equal(expected.Error()))
			})

			It("errors when invalid tag in changelog directory", func() {
				repoClient.EXPECT().
					FindLatestReleaseTagIncudingPrerelease(ctx).
					Return(tag, nil)
				code.EXPECT().
					ListFiles(ctx, changelogutils.ChangelogDirectory).
					Return([]os.FileInfo{getChangelogDir("invalid-tag")}, nil)

				expected := changelogutils.InvalidChangelogSubdirectoryNameError("invalid-tag")
				file, err := validator.ValidateChangelog(ctx)
				Expect(file).To(BeNil())
				Expect(err.Error()).To(Equal(expected.Error()))
			})

			It("errors when no new version", func() {
				repoClient.EXPECT().
					FindLatestReleaseTagIncudingPrerelease(ctx).
					Return(tag, nil)
				code.EXPECT().
					ListFiles(ctx, changelogutils.ChangelogDirectory).
					Return([]os.FileInfo{getChangelogDir(tag)}, nil)

				expected := changelogutils.NoNewVersionsFoundError(tag)
				file, err := validator.ValidateChangelog(ctx)
				Expect(file).To(BeNil())
				Expect(err.Error()).To(Equal(expected.Error()))
			})

			It("errors when no too many new versions", func() {
				repoClient.EXPECT().
					FindLatestReleaseTagIncudingPrerelease(ctx).
					Return(tag, nil)
				code.EXPECT().
					ListFiles(ctx, changelogutils.ChangelogDirectory).
					Return([]os.FileInfo{getChangelogDir("v0.5.2"), getChangelogDir("v0.5.3")}, nil)

				expected := changelogutils.MultipleNewVersionsFoundError(tag, "v0.5.2", "v0.5.3")
				file, err := validator.ValidateChangelog(ctx)
				Expect(file).To(BeNil())
				Expect(err.Error()).To(Equal(expected.Error()))
			})

			It("errors when added changelog to old version", func() {
				repoClient.EXPECT().
					FindLatestReleaseTagIncudingPrerelease(ctx).
					Return(tag, nil)
				code.EXPECT().
					ListFiles(ctx, changelogutils.ChangelogDirectory).
					Return([]os.FileInfo{getChangelogDir(nextTag)}, nil)
				code.EXPECT().
					ListFiles(ctx, filepath.Join(changelogutils.ChangelogDirectory, nextTag)).
					Return([]os.FileInfo{&mockFileInfo{name: filename3, isDir: false}}, nil)
				code.EXPECT().
					GetFileContents(ctx, path3).
					Return([]byte(validChangelog2), nil)
				expected := changelogutils.AddedChangelogInOldVersionError(nextTag)
				file, err := validator.ValidateChangelog(ctx)
				Expect(file).To(BeNil())
				Expect(err.Error()).To(Equal(expected.Error()))
			})
		})



	})
})
