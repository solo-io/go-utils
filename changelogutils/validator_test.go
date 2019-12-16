package changelogutils_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/changelogutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/githubutils"
)

var _ = Describe("github utils", func() {

	const (
		base = "base"
		sha  = "sha"
	)

	var (
		ctrl       *gomock.Controller
		repoClient *MockRepoClient
		code       *MockMountedRepo
		validator  changelogutils.ChangelogValidator
		ctx        = context.Background()
		nestedErr  = errors.Errorf("")
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
			tag       = "v0.5.1"
			nextTag   = "v0.5.2"
		)

		var (
			path1          = filepath.Join(changelogutils.ChangelogDirectory, tag, filename1)
			path2          = filepath.Join(changelogutils.ChangelogDirectory, tag, filename2)
			added          = githubutils.COMMIT_FILE_STATUS_ADDED
			unexpectedFile = mockFileInfo{isDir: false, name: "unexpected"}

			path3 = filepath.Join(changelogutils.ChangelogDirectory, nextTag, filename3)
		)

		relaxedValidationSettingsExists := func() {
			repoClient.EXPECT().FileExists(ctx, sha, changelogutils.GetValidationSettingsPath()).Return(true, nil)
			code.EXPECT().GetFileContents(ctx, changelogutils.GetValidationSettingsPath()).Return([]byte(validationYaml), nil)
		}

		noValidationSettingsExist := func() {
			repoClient.EXPECT().FileExists(ctx, sha, changelogutils.GetValidationSettingsPath()).Return(false, nil)
		}

		getChangelogDir := func(tag string) os.FileInfo {
			return &mockFileInfo{name: tag, isDir: true}
		}

		validateVersionBump := func(lastTag, nextTag, contents string, expectFailure bool) {

			nextTagFile := filepath.Join(changelogutils.ChangelogDirectory, nextTag, filename1)
			cc := github.CommitsComparison{Files: []github.CommitFile{{Filename: &nextTagFile, Status: &added}}}

			repoClient.EXPECT().
				CompareCommits(ctx, base, sha).
				Return(&cc, nil)

			code.EXPECT().
				GetFileContents(ctx, nextTagFile).
				Return([]byte(contents), nil).Times(2)

			repoClient.EXPECT().
				FindLatestTagIncludingPrereleaseBeforeSha(ctx, base).
				Return(lastTag, nil)

			code.EXPECT().
				ListFiles(ctx, changelogutils.ChangelogDirectory).
				Return([]os.FileInfo{getChangelogDir(nextTag)}, nil)

			code.EXPECT().
				ListFiles(ctx, filepath.Join(changelogutils.ChangelogDirectory, nextTag)).
				Return([]os.FileInfo{&mockFileInfo{name: filename1, isDir: false}}, nil)

			file, err := validator.ValidateChangelog(ctx)

			if expectFailure {
				Expect(err).To(HaveOccurred())
				Expect(file).To(BeNil())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(file).NotTo(BeNil())
			}
		}

		BeforeEach(func() {
			// so should check returns true
			repoClient.EXPECT().
				DirectoryExists(ctx, changelogutils.MasterBranch, changelogutils.ChangelogDirectory).
				Return(true, nil)
			code.EXPECT().GetSha().Return(sha).AnyTimes()
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
					FindLatestTagIncludingPrereleaseBeforeSha(ctx, base).
					Return("", nestedErr)

				expected := changelogutils.ListReleasesError(nestedErr)
				file, err := validator.ValidateChangelog(ctx)
				Expect(file).To(BeNil())
				Expect(err.Error()).To(Equal(expected.Error()))
			})

			It("errors when unexpected file in changelog directory", func() {
				repoClient.EXPECT().
					FindLatestTagIncludingPrereleaseBeforeSha(ctx, base).
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
					FindLatestTagIncludingPrereleaseBeforeSha(ctx, base).
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
					FindLatestTagIncludingPrereleaseBeforeSha(ctx, base).
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
					FindLatestTagIncludingPrereleaseBeforeSha(ctx, base).
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
					FindLatestTagIncludingPrereleaseBeforeSha(ctx, base).
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
				repoClient.EXPECT().
					FileExists(ctx, sha, changelogutils.GetValidationSettingsPath()).
					Return(false, nil)
				expected := changelogutils.AddedChangelogInOldVersionError(nextTag)
				file, err := validator.ValidateChangelog(ctx)
				Expect(file).To(BeNil())
				Expect(err.Error()).To(Equal(expected.Error()))
			})

		})

		Context("incrementing versions", func() {

			BeforeEach(func() {
				noValidationSettingsExist()
			})

			Context("major version zero 0.y.z", func() {

				It("works on patch version bump", func() {
					file1 := github.CommitFile{Filename: &path1, Status: &added}
					cc := github.CommitsComparison{Files: []github.CommitFile{file1}}
					repoClient.EXPECT().
						CompareCommits(ctx, base, sha).
						Return(&cc, nil)
					code.EXPECT().
						GetFileContents(ctx, path1).
						Return([]byte(validChangelog1), nil).Times(2)
					repoClient.EXPECT().
						FindLatestTagIncludingPrereleaseBeforeSha(ctx, base).
						Return("v0.5.0", nil)
					code.EXPECT().
						ListFiles(ctx, changelogutils.ChangelogDirectory).
						Return([]os.FileInfo{getChangelogDir(tag)}, nil)
					code.EXPECT().
						ListFiles(ctx, filepath.Join(changelogutils.ChangelogDirectory, tag)).
						Return([]os.FileInfo{&mockFileInfo{name: filename1, isDir: false}}, nil)
					file, err := validator.ValidateChangelog(ctx)
					Expect(file).NotTo(BeNil())
					Expect(err).To(BeNil())
				})

				It("errors when not incrementing major version", func() {
					file1 := github.CommitFile{Filename: &path1, Status: &added}
					cc := github.CommitsComparison{Files: []github.CommitFile{file1}}
					repoClient.EXPECT().
						CompareCommits(ctx, base, sha).
						Return(&cc, nil)
					code.EXPECT().
						GetFileContents(ctx, path1).
						Return([]byte(validBreakingChangelog), nil).Times(2)
					repoClient.EXPECT().
						FindLatestTagIncludingPrereleaseBeforeSha(ctx, base).
						Return("v0.5.0", nil)
					code.EXPECT().
						ListFiles(ctx, changelogutils.ChangelogDirectory).
						Return([]os.FileInfo{getChangelogDir(tag)}, nil)
					code.EXPECT().
						ListFiles(ctx, filepath.Join(changelogutils.ChangelogDirectory, tag)).
						Return([]os.FileInfo{&mockFileInfo{name: filename1, isDir: false}}, nil)

					expected := changelogutils.UnexpectedProposedVersionError("v0.6.0", tag)
					file, err := validator.ValidateChangelog(ctx)
					Expect(file).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(expected.Error()))
				})

				It("works when incrementing major version", func() {
					path := filepath.Join(changelogutils.ChangelogDirectory, "v0.6.0", filename1)
					file1 := github.CommitFile{Filename: &path, Status: &added}
					cc := github.CommitsComparison{Files: []github.CommitFile{file1}}
					repoClient.EXPECT().
						CompareCommits(ctx, base, sha).
						Return(&cc, nil)
					code.EXPECT().
						GetFileContents(ctx, path).
						Return([]byte(validBreakingChangelog), nil).Times(2)
					repoClient.EXPECT().
						FindLatestTagIncludingPrereleaseBeforeSha(ctx, base).
						Return("v0.5.0", nil)
					code.EXPECT().
						ListFiles(ctx, changelogutils.ChangelogDirectory).
						Return([]os.FileInfo{getChangelogDir("v0.6.0")}, nil)
					code.EXPECT().
						ListFiles(ctx, filepath.Join(changelogutils.ChangelogDirectory, "v0.6.0")).
						Return([]os.FileInfo{&mockFileInfo{name: filename1, isDir: false}}, nil)

					file, err := validator.ValidateChangelog(ctx)
					Expect(err).To(BeNil())
					Expect(file).NotTo(BeNil())
				})

			})

			Context("major version is 1.y.z", func() {

				DescribeTable("correctly enforces version bump rules",
					validateVersionBump,
					Entry("breaking change with patch bump", "v1.0.0", "v1.0.1", validBreakingChangelog, true),
					Entry("breaking change with minor bump", "v1.0.0", "v1.1.0", validBreakingChangelog, true),
					Entry("breaking change with major bump", "v1.0.0", "v2.0.0", validBreakingChangelog, false),
					Entry("new feature with patch bump", "v1.0.0", "v1.0.1", validNewFeatureChangelog, true),
					Entry("new feature with minor bump", "v1.0.0", "v1.1.0", validNewFeatureChangelog, false),
					Entry("new feature with major bump", "v1.0.0", "v2.0.0", validNewFeatureChangelog, true),
					Entry("non-breaking with patch bump", "v1.0.0", "v1.0.1", validNonBreakingNorNewFeatureChangelog, false),
					Entry("non-breaking with minor bump", "v1.0.0", "v1.1.0", validNonBreakingNorNewFeatureChangelog, true),
					Entry("non-breaking with major bump", "v1.0.0", "v2.0.0", validNonBreakingNorNewFeatureChangelog, true),
				)
			})

			Context("moving from 0.x to 1.x", func() {

				It("works for stable api release", func() {
					path := filepath.Join(changelogutils.ChangelogDirectory, "v1.0.0", filename1)
					file1 := github.CommitFile{Filename: &path, Status: &added}
					cc := github.CommitsComparison{Files: []github.CommitFile{file1}}
					repoClient.EXPECT().
						CompareCommits(ctx, base, sha).
						Return(&cc, nil)
					code.EXPECT().
						GetFileContents(ctx, path).
						Return([]byte(validStableReleaseChangelog), nil).Times(2)
					repoClient.EXPECT().
						FindLatestTagIncludingPrereleaseBeforeSha(ctx, base).
						Return("v0.5.0", nil)
					code.EXPECT().
						ListFiles(ctx, changelogutils.ChangelogDirectory).
						Return([]os.FileInfo{getChangelogDir("v1.0.0")}, nil)
					code.EXPECT().
						ListFiles(ctx, filepath.Join(changelogutils.ChangelogDirectory, "v1.0.0")).
						Return([]os.FileInfo{&mockFileInfo{name: filename1, isDir: false}}, nil)

					file, err := validator.ValidateChangelog(ctx)
					Expect(err).To(BeNil())
					Expect(file).NotTo(BeNil())
				})

				It("errors when not incrementing for stable api release", func() {
					path := filepath.Join(changelogutils.ChangelogDirectory, nextTag, filename1)
					file1 := github.CommitFile{Filename: &path, Status: &added}
					cc := github.CommitsComparison{Files: []github.CommitFile{file1}}
					repoClient.EXPECT().
						CompareCommits(ctx, base, sha).
						Return(&cc, nil)
					code.EXPECT().
						GetFileContents(ctx, path).
						Return([]byte(validStableReleaseChangelog), nil).Times(2)
					repoClient.EXPECT().
						FindLatestTagIncludingPrereleaseBeforeSha(ctx, base).
						Return("v0.5.0", nil)
					code.EXPECT().
						ListFiles(ctx, changelogutils.ChangelogDirectory).
						Return([]os.FileInfo{getChangelogDir(nextTag)}, nil)
					code.EXPECT().
						ListFiles(ctx, filepath.Join(changelogutils.ChangelogDirectory, nextTag)).
						Return([]os.FileInfo{&mockFileInfo{name: filename1, isDir: false}}, nil)

					expected := changelogutils.InvalidUseOfStableApiError(nextTag)
					file, err := validator.ValidateChangelog(ctx)
					Expect(err.Error()).To(Equal(expected.Error()))
					Expect(file).To(BeNil())
				})
			})
		})

		Context("incrementing versions with relaxed validation", func() {

			BeforeEach(func() {
				relaxedValidationSettingsExists()
			})

			Context("major version zero 0.y.z", func() {

				It("allows not incrementing major version", func() {
					file1 := github.CommitFile{Filename: &path1, Status: &added}
					cc := github.CommitsComparison{Files: []github.CommitFile{file1}}
					repoClient.EXPECT().
						CompareCommits(ctx, base, sha).
						Return(&cc, nil)
					code.EXPECT().
						GetFileContents(ctx, path1).
						Return([]byte(validBreakingChangelog), nil).Times(2)
					repoClient.EXPECT().
						FindLatestTagIncludingPrereleaseBeforeSha(ctx, base).
						Return("v0.5.0", nil)
					code.EXPECT().
						ListFiles(ctx, changelogutils.ChangelogDirectory).
						Return([]os.FileInfo{getChangelogDir(tag)}, nil)
					code.EXPECT().
						ListFiles(ctx, filepath.Join(changelogutils.ChangelogDirectory, tag)).
						Return([]os.FileInfo{&mockFileInfo{name: filename1, isDir: false}}, nil)

					file, err := validator.ValidateChangelog(ctx)
					Expect(file).NotTo(BeNil())
					Expect(err).To(BeNil())
				})

				It("works when incrementing major version", func() {
					path := filepath.Join(changelogutils.ChangelogDirectory, "v0.6.0", filename1)
					file1 := github.CommitFile{Filename: &path, Status: &added}
					cc := github.CommitsComparison{Files: []github.CommitFile{file1}}
					repoClient.EXPECT().
						CompareCommits(ctx, base, sha).
						Return(&cc, nil)
					code.EXPECT().
						GetFileContents(ctx, path).
						Return([]byte(validBreakingChangelog), nil).Times(2)
					repoClient.EXPECT().
						FindLatestTagIncludingPrereleaseBeforeSha(ctx, base).
						Return("v0.5.0", nil)
					code.EXPECT().
						ListFiles(ctx, changelogutils.ChangelogDirectory).
						Return([]os.FileInfo{getChangelogDir("v0.6.0")}, nil)
					code.EXPECT().
						ListFiles(ctx, filepath.Join(changelogutils.ChangelogDirectory, "v0.6.0")).
						Return([]os.FileInfo{&mockFileInfo{name: filename1, isDir: false}}, nil)

					file, err := validator.ValidateChangelog(ctx)
					Expect(err).To(BeNil())
					Expect(file).NotTo(BeNil())
				})

			})

			Context("major version is 1.y.z", func() {

				DescribeTable("doesn't enforce version bump rules",
					validateVersionBump,
					Entry("breaking change with patch bump", "v1.0.0", "v1.0.1", validBreakingChangelog, false),
					Entry("breaking change with minor bump", "v1.0.0", "v1.1.0", validBreakingChangelog, false),
					Entry("breaking change with major bump", "v1.0.0", "v2.0.0", validBreakingChangelog, false),
					Entry("new feature with patch bump", "v1.0.0", "v1.0.1", validNewFeatureChangelog, false),
					Entry("new feature with minor bump", "v1.0.0", "v1.1.0", validNewFeatureChangelog, false),
					Entry("new feature with major bump", "v1.0.0", "v2.0.0", validNewFeatureChangelog, false),
					Entry("non-breaking with patch bump", "v1.0.0", "v1.0.1", validNonBreakingNorNewFeatureChangelog, false),
					Entry("non-breaking with minor bump", "v1.0.0", "v1.1.0", validNonBreakingNorNewFeatureChangelog, false),
					Entry("non-breaking with major bump", "v1.0.0", "v2.0.0", validNonBreakingNorNewFeatureChangelog, false),
					Entry("stable release that skips to a future version", "v1.0.0-rc6", "v1.2.0", validNonBreakingNorNewFeatureChangelog, false),
				)
			})

			Context("moving from 0.x to 1.x", func() {

				It("works for stable api release", func() {
					path := filepath.Join(changelogutils.ChangelogDirectory, "v1.0.0", filename1)
					file1 := github.CommitFile{Filename: &path, Status: &added}
					cc := github.CommitsComparison{Files: []github.CommitFile{file1}}
					repoClient.EXPECT().
						CompareCommits(ctx, base, sha).
						Return(&cc, nil)
					code.EXPECT().
						GetFileContents(ctx, path).
						Return([]byte(validStableReleaseChangelog), nil).Times(2)
					repoClient.EXPECT().
						FindLatestTagIncludingPrereleaseBeforeSha(ctx, base).
						Return("v0.5.0", nil)
					code.EXPECT().
						ListFiles(ctx, changelogutils.ChangelogDirectory).
						Return([]os.FileInfo{getChangelogDir("v1.0.0")}, nil)
					code.EXPECT().
						ListFiles(ctx, filepath.Join(changelogutils.ChangelogDirectory, "v1.0.0")).
						Return([]os.FileInfo{&mockFileInfo{name: filename1, isDir: false}}, nil)

					file, err := validator.ValidateChangelog(ctx)
					Expect(err).To(BeNil())
					Expect(file).NotTo(BeNil())
				})

				It("errors when not incrementing for stable api release", func() {
					path := filepath.Join(changelogutils.ChangelogDirectory, nextTag, filename1)
					file1 := github.CommitFile{Filename: &path, Status: &added}
					cc := github.CommitsComparison{Files: []github.CommitFile{file1}}
					repoClient.EXPECT().
						CompareCommits(ctx, base, sha).
						Return(&cc, nil)
					code.EXPECT().
						GetFileContents(ctx, path).
						Return([]byte(validStableReleaseChangelog), nil).Times(2)
					repoClient.EXPECT().
						FindLatestTagIncludingPrereleaseBeforeSha(ctx, base).
						Return("v0.5.0", nil)
					code.EXPECT().
						ListFiles(ctx, changelogutils.ChangelogDirectory).
						Return([]os.FileInfo{getChangelogDir(nextTag)}, nil)
					code.EXPECT().
						ListFiles(ctx, filepath.Join(changelogutils.ChangelogDirectory, nextTag)).
						Return([]os.FileInfo{&mockFileInfo{name: filename1, isDir: false}}, nil)

					expected := changelogutils.InvalidUseOfStableApiError(nextTag)
					file, err := validator.ValidateChangelog(ctx)
					Expect(err.Error()).To(Equal(expected.Error()))
					Expect(file).To(BeNil())
				})
			})
		})

		Context("rc workflow", func() {

			rcWorkflow := func(lastTag, nextTag, contents string, settingsFunc func()) {
				path := filepath.Join(changelogutils.ChangelogDirectory, nextTag, filename1)
				file1 := github.CommitFile{Filename: &path, Status: &added}
				cc := github.CommitsComparison{Files: []github.CommitFile{file1}}
				repoClient.EXPECT().
					CompareCommits(ctx, base, sha).
					Return(&cc, nil)
				code.EXPECT().
					GetFileContents(ctx, path).
					Return([]byte(validBreakingChangelog), nil)
				repoClient.EXPECT().
					FindLatestTagIncludingPrereleaseBeforeSha(ctx, base).
					Return(lastTag, nil)
				code.EXPECT().
					ListFiles(ctx, changelogutils.ChangelogDirectory).
					Return([]os.FileInfo{getChangelogDir(nextTag)}, nil)
				code.EXPECT().
					ListFiles(ctx, filepath.Join(changelogutils.ChangelogDirectory, nextTag)).
					Return([]os.FileInfo{&mockFileInfo{name: filename1, isDir: false}}, nil)
				code.EXPECT().
					GetFileContents(ctx, path).
					Return([]byte(contents), nil)

				settingsFunc()

				file, err := validator.ValidateChangelog(ctx)
				Expect(err).To(BeNil())
				Expect(file).NotTo(BeNil())
			}

			DescribeTable("rc workflow cases",
				rcWorkflow,
				Entry("initial rc", "v0.20.5", "v1.0.0-rc1", validBreakingChangelog, noValidationSettingsExist),
				Entry("initial rc relaxed", "v0.20.5", "v1.0.0-rc1", validBreakingChangelog, relaxedValidationSettingsExists),
				Entry("incrementing rc", "v1.0.0-rc1", "v1.0.0-rc2", validBreakingChangelog, noValidationSettingsExist),
				Entry("incrementing rc", "v1.0.0-rc1", "v1.0.0-rc2", validBreakingChangelog, relaxedValidationSettingsExists),
				Entry("stable release after rc for 1.0", "v1.0.0-rc2", "v1.0.0", validStableReleaseChangelog, noValidationSettingsExist),
				Entry("stable release after rc for 1.0", "v1.0.0-rc2", "v1.0.0", validStableReleaseChangelog, relaxedValidationSettingsExists),
				Entry("stable release after rc for 1.1", "v1.1.0-rc2", "v1.1.0", validStableReleaseChangelog, noValidationSettingsExist),
				Entry("stable release after rc for 1.1", "v1.1.0-rc2", "v1.1.0", validStableReleaseChangelog, relaxedValidationSettingsExists),
				Entry("initial rc after for 1.1", "v1.0.0", "v1.1.0-rc1", validNewFeatureChangelog, noValidationSettingsExist),
				Entry("initial rc after for 1.1", "v1.0.0", "v1.1.0-rc1", validNewFeatureChangelog, relaxedValidationSettingsExists))
		})
	})
})

const (
	validationYaml = `
relaxSemverValidation: true
`
)
