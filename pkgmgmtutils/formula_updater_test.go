package pkgmgmtutils_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/pkgmgmtutils/formula_updater_types"
	mock_formula_updater_types "github.com/solo-io/go-utils/pkgmgmtutils/formula_updater_types/mocks"
	"github.com/solo-io/go-utils/versionutils"

	"github.com/solo-io/go-utils/pkgmgmtutils"
)

var _ = Describe("FormulaUpdater", func() {
	var (
		ctx               = context.Background()
		ctrl              *gomock.Controller
		gitClient         *mock_formula_updater_types.MockGitClient
		remoteShaGetter   *mock_formula_updater_types.MockRemoteShaGetter
		epochTime         = int64(0)
		epochTimeProvider = func() int64 {
			epochTime++
			return epochTime
		}
		localCloneChangePusher *mock_formula_updater_types.MockChangePusher
		remoteChangePusher     *mock_formula_updater_types.MockChangePusher
		formulaUpdater         *pkgmgmtutils.FormulaUpdater
		mustParseVersion       = func(version string) *versionutils.Version {
			parsedVersion, err := versionutils.ParseVersion(version)
			Expect(err).NotTo(HaveOccurred())
			return parsedVersion
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())

		epochTime = 0
		gitClient = mock_formula_updater_types.NewMockGitClient(ctrl)
		remoteShaGetter = mock_formula_updater_types.NewMockRemoteShaGetter(ctrl)
		localCloneChangePusher = mock_formula_updater_types.NewMockChangePusher(ctrl)
		remoteChangePusher = mock_formula_updater_types.NewMockChangePusher(ctrl)
		formulaUpdater = pkgmgmtutils.NewFormulaUpdater(
			gitClient,
			remoteShaGetter,
			epochTimeProvider,
			localCloneChangePusher,
			remoteChangePusher,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	When("no formula options are provided", func() {
		It("does nothing", func() {
			status, err := formulaUpdater.Update(ctx, mustParseVersion("v1.0.0"), "solo-io", "test-repo", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(BeEmpty())
		})
	})

	When("updating a formula using the Gloo use case", func() {
		Context("and we are releasing a stable version", func() {
			It("works", func() {
				repoOwner := "solo-io"
				repoName := "gloo"
				formulaOptionsList := []*formula_updater_types.FormulaOptions{
					{
						Name:           "homebrew-tap/glooctl",
						FormulaName:    "glooctl",
						Path:           "Formula/glooctl.rb",
						RepoOwner:      repoOwner,      // Make change in this repo
						RepoName:       "homebrew-tap", // assumes this repo is forked from PRRepoOwner
						PRRepoOwner:    repoOwner,      // Make PR to this repo
						PRRepoName:     "homebrew-tap",
						PRBranch:       "master",
						PRDescription:  "",
						PRCommitName:   "Solo-io Bot",
						PRCommitEmail:  "bot@solo.io",
						VersionRegex:   `version\s*"([0-9.]+)"`,
						DarwinShaRegex: `url\s*".*-darwin.*\W*sha256\s*"(.*)"`,
						LinuxShaRegex:  `url\s*".*-linux.*\W*sha256\s*"(.*)"`,
					},
					{
						Name:            "fish-food/glooctl",
						FormulaName:     "glooctl",
						Path:            "Food/glooctl.lua",
						RepoOwner:       repoOwner,
						RepoName:        "fish-food",
						PRRepoOwner:     "fishworks",
						PRRepoName:      "fish-food",
						PRBranch:        "master",
						PRDescription:   "",
						PRCommitName:    "Solo-io Bot",
						PRCommitEmail:   "bot@solo.io",
						VersionRegex:    `version\s*=\s*"([0-9.]+)"`,
						DarwinShaRegex:  `os\s*=\s*"darwin",\W*.*\W*.*\W*.*\W*sha256\s*=\s*"(.*)",`,
						LinuxShaRegex:   `os\s*=\s*"linux",\W*.*\W*.*\W*.*\W*sha256\s*=\s*"(.*)",`,
						WindowsShaRegex: `os\s*=\s*"windows",\W*.*\W*.*\W*.*\W*sha256\s*=\s*"(.*)",`,
					},
					{
						Name:            "homebrew-core/glooctl",
						FormulaName:     "glooctl",
						Path:            "Formula/glooctl.rb",
						RepoOwner:       repoOwner,
						RepoName:        "homebrew-core",
						PRRepoOwner:     "homebrew",
						PRRepoName:      "homebrew-core",
						PRBranch:        "master",
						PRDescription:   "Created by Solo-io Bot",
						PRCommitName:    "Solo-io Bot",
						PRCommitEmail:   "bot@solo.io",
						VersionRegex:    `:tag\s*=>\s*"v([0-9.]+)",`,
						VersionShaRegex: `:revision\s*=>\s*"(.*)"`,
					},
				}
				version := "v1.6.9"
				gitSha := "my-really-long-sha256-string"
				commitMessage := "glooctl " + version[1:]
				branch1Name := fmt.Sprintf("glooctl-%s-%d", version[1:], 1)
				branch2Name := fmt.Sprintf("glooctl-%s-%d", version[1:], 2)
				branch3Name := fmt.Sprintf("glooctl-%s-%d", version[1:], 3)

				gitClient.EXPECT().
					GetRefSha(ctx, repoOwner, repoName, "refs/tags/"+version).
					Return(gitSha, nil)
				gitClient.EXPECT().
					GetReleaseAssetsByTag(ctx, repoOwner, repoName, version).
					Return([]formula_updater_types.ReleaseAsset{
						{
							Name:               "glooctl-darwin",
							BrowserDownloadUrl: "darwin-download-url",
						},
						{
							Name:               "glooctl-linux",
							BrowserDownloadUrl: "linux-download-url",
						},
						{
							Name:               "glooctl-windows",
							BrowserDownloadUrl: "windows-download-url",
						},
						{
							Name:               "glooctl-darwin.sha256",
							BrowserDownloadUrl: "darwin-checksum-download-url",
						},
						{
							Name:               "glooctl-linux.sha256",
							BrowserDownloadUrl: "linux-checksum-download-url",
						},
						{
							Name:               "glooctl-windows.sha256",
							BrowserDownloadUrl: "windows-checksum-download-url",
						},
					}, nil)
				remoteShaGetter.EXPECT().
					GetShaFromUrl("darwin-checksum-download-url").
					Return("darwin-cli-sha", nil)
				remoteShaGetter.EXPECT().
					GetShaFromUrl("linux-checksum-download-url").
					Return("linux-cli-sha", nil)
				remoteShaGetter.EXPECT().
					GetShaFromUrl("windows-checksum-download-url").
					Return("windows-cli-sha", nil)

				remoteChangePusher.EXPECT().
					UpdateAndPush(ctx, version[1:], gitSha, branch1Name, commitMessage, &formula_updater_types.PerPlatformSha256{
						DarwinSha:  "darwin-cli-sha",
						LinuxSha:   "linux-cli-sha",
						WindowsSha: "windows-cli-sha",
					}, formulaOptionsList[0]).
					Return(nil)
				gitClient.EXPECT().
					CreatePullRequest(ctx, formulaOptionsList[0], commitMessage, branch1Name).
					Return(nil)

				localCloneChangePusher.EXPECT().
					UpdateAndPush(ctx, version[1:], gitSha, branch2Name, commitMessage, &formula_updater_types.PerPlatformSha256{
						DarwinSha:  "darwin-cli-sha",
						LinuxSha:   "linux-cli-sha",
						WindowsSha: "windows-cli-sha",
					}, formulaOptionsList[1]).
					Return(nil)
				gitClient.EXPECT().
					CreatePullRequest(ctx, formulaOptionsList[1], commitMessage, branch2Name).
					Return(nil)

				localCloneChangePusher.EXPECT().
					UpdateAndPush(ctx, version[1:], gitSha, branch3Name, commitMessage, &formula_updater_types.PerPlatformSha256{
						DarwinSha:  "darwin-cli-sha",
						LinuxSha:   "linux-cli-sha",
						WindowsSha: "windows-cli-sha",
					}, formulaOptionsList[2]).
					Return(nil)
				gitClient.EXPECT().
					CreatePullRequest(ctx, formulaOptionsList[2], commitMessage, branch3Name).
					Return(nil)

				statuses, err := formulaUpdater.Update(ctx, mustParseVersion(version), repoOwner, repoName, formulaOptionsList)
				Expect(err).NotTo(HaveOccurred())
				Expect(statuses).To(HaveLen(3))

				for _, status := range statuses {
					Expect(status.Err).NotTo(HaveOccurred())
					Expect(status.Updated).To(BeTrue())
				}
			})
		})

		Context("and we are releasing an unstable version", func() {
			It("does not open PRs by default", func() {
				repoOwner := "solo-io"
				repoName := "gloo"
				formulaOptionsList := []*formula_updater_types.FormulaOptions{
					{
						Name:                   "homebrew-tap/glooctl",
						FormulaName:            "glooctl",
						Path:                   "Formula/glooctl.rb",
						RepoOwner:              repoOwner,      // Make change in this repo
						RepoName:               "homebrew-tap", // assumes this repo is forked from PRRepoOwner
						PRRepoOwner:            repoOwner,      // Make PR to this repo
						PRRepoName:             "homebrew-tap",
						PRBranch:               "master",
						PRDescription:          "",
						PRCommitName:           "Solo-io Bot",
						PRCommitEmail:          "bot@solo.io",
						VersionRegex:           `version\s*"([0-9.]+)"`,
						DarwinShaRegex:         `url\s*".*-darwin.*\W*sha256\s*"(.*)"`,
						LinuxShaRegex:          `url\s*".*-linux.*\W*sha256\s*"(.*)"`,
						PublishUnstableVersion: true,
					},
					{
						Name:                   "fish-food/glooctl",
						FormulaName:            "glooctl",
						Path:                   "Food/glooctl.lua",
						RepoOwner:              repoOwner,
						RepoName:               "fish-food",
						PRRepoOwner:            "fishworks",
						PRRepoName:             "fish-food",
						PRBranch:               "master",
						PRDescription:          "",
						PRCommitName:           "Solo-io Bot",
						PRCommitEmail:          "bot@solo.io",
						VersionRegex:           `version\s*=\s*"([0-9.]+)"`,
						DarwinShaRegex:         `os\s*=\s*"darwin",\W*.*\W*.*\W*.*\W*sha256\s*=\s*"(.*)",`,
						LinuxShaRegex:          `os\s*=\s*"linux",\W*.*\W*.*\W*.*\W*sha256\s*=\s*"(.*)",`,
						WindowsShaRegex:        `os\s*=\s*"windows",\W*.*\W*.*\W*.*\W*sha256\s*=\s*"(.*)",`,
						PublishUnstableVersion: true,
					},
					{
						Name:            "homebrew-core/glooctl",
						FormulaName:     "glooctl",
						Path:            "Formula/glooctl.rb",
						RepoOwner:       repoOwner,
						RepoName:        "homebrew-core",
						PRRepoOwner:     "homebrew",
						PRRepoName:      "homebrew-core",
						PRBranch:        "master",
						PRDescription:   "Created by Solo-io Bot",
						PRCommitName:    "Solo-io Bot",
						PRCommitEmail:   "bot@solo.io",
						VersionRegex:    `:tag\s*=>\s*"v([0-9.]+)",`,
						VersionShaRegex: `:revision\s*=>\s*"(.*)"`,
					},
				}
				version := "v1.6.9-beta420"
				gitSha := "my-really-long-sha256-string"
				commitMessage := "glooctl " + version[1:]
				branch1Name := fmt.Sprintf("glooctl-%s-%d", version[1:], 1)
				branch2Name := fmt.Sprintf("glooctl-%s-%d", version[1:], 2)

				gitClient.EXPECT().
					GetRefSha(ctx, repoOwner, repoName, "refs/tags/"+version).
					Return(gitSha, nil)
				gitClient.EXPECT().
					GetReleaseAssetsByTag(ctx, repoOwner, repoName, version).
					Return([]formula_updater_types.ReleaseAsset{
						{
							Name:               "glooctl-darwin",
							BrowserDownloadUrl: "darwin-download-url",
						},
						{
							Name:               "glooctl-linux",
							BrowserDownloadUrl: "linux-download-url",
						},
						{
							Name:               "glooctl-windows",
							BrowserDownloadUrl: "windows-download-url",
						},
						{
							Name:               "glooctl-darwin.sha256",
							BrowserDownloadUrl: "darwin-checksum-download-url",
						},
						{
							Name:               "glooctl-linux.sha256",
							BrowserDownloadUrl: "linux-checksum-download-url",
						},
						{
							Name:               "glooctl-windows.sha256",
							BrowserDownloadUrl: "windows-checksum-download-url",
						},
					}, nil)
				remoteShaGetter.EXPECT().
					GetShaFromUrl("darwin-checksum-download-url").
					Return("darwin-cli-sha", nil)
				remoteShaGetter.EXPECT().
					GetShaFromUrl("linux-checksum-download-url").
					Return("linux-cli-sha", nil)
				remoteShaGetter.EXPECT().
					GetShaFromUrl("windows-checksum-download-url").
					Return("windows-cli-sha", nil)

				remoteChangePusher.EXPECT().
					UpdateAndPush(ctx, version[1:], gitSha, branch1Name, commitMessage, &formula_updater_types.PerPlatformSha256{
						DarwinSha:  "darwin-cli-sha",
						LinuxSha:   "linux-cli-sha",
						WindowsSha: "windows-cli-sha",
					}, formulaOptionsList[0]).
					Return(nil)
				gitClient.EXPECT().
					CreatePullRequest(ctx, formulaOptionsList[0], commitMessage, branch1Name).
					Return(nil)

				localCloneChangePusher.EXPECT().
					UpdateAndPush(ctx, version[1:], gitSha, branch2Name, commitMessage, &formula_updater_types.PerPlatformSha256{
						DarwinSha:  "darwin-cli-sha",
						LinuxSha:   "linux-cli-sha",
						WindowsSha: "windows-cli-sha",
					}, formulaOptionsList[1]).
					Return(nil)
				gitClient.EXPECT().
					CreatePullRequest(ctx, formulaOptionsList[1], commitMessage, branch2Name).
					Return(nil)

				// NOTE: not setting up expectations for the homebrew-core cask here

				statuses, err := formulaUpdater.Update(ctx, mustParseVersion(version), repoOwner, repoName, formulaOptionsList)
				Expect(err).NotTo(HaveOccurred())
				Expect(statuses).To(HaveLen(3))

				for _, status := range statuses {
					Expect(status.Err).NotTo(HaveOccurred())
					Expect(status.Updated).To(BeTrue())
				}
			})
		})
	})
})
