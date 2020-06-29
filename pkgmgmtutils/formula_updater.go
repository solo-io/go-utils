package pkgmgmtutils

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/pkgmgmtutils/formula_updater_types"
	"github.com/solo-io/go-utils/pkgmgmtutils/internal"
	"github.com/solo-io/go-utils/versionutils"
)

var (
	ErrAlreadyUpdated = eris.New("pkgmgmtutils: formula already updated")
	ErrNoSha256sFound = eris.New("pkgmgmtutils: did not find any sha256 data")
)

func NewFormulaUpdater(
	gitClient formula_updater_types.GitClient,
	remoteShaGetter formula_updater_types.RemoteShaGetter,
	currentEpochTimeProvider func() int64,
	localCloneChangePusher formula_updater_types.ChangePusher,
	remoteChangePusher formula_updater_types.ChangePusher,
) *FormulaUpdater {
	return &FormulaUpdater{
		gitClient:                gitClient,
		remoteShaGetter:          remoteShaGetter,
		currentEpochTimeProvider: currentEpochTimeProvider,
		localCloneChangePusher:   localCloneChangePusher,
		remoteChangePusher:       remoteChangePusher,
	}
}

type FormulaUpdater struct {
	gitClient                formula_updater_types.GitClient
	remoteShaGetter          formula_updater_types.RemoteShaGetter
	currentEpochTimeProvider func() int64
	localCloneChangePusher   formula_updater_types.ChangePusher
	remoteChangePusher       formula_updater_types.ChangePusher
}

func NewFormulaUpdaterWithDefaults(ctx context.Context) (*FormulaUpdater, error) {
	client, err := githubutils.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	return NewFormulaUpdater(
		internal.NewGitClient(client),
		internal.NewRemoteShaGetter(),
		func() int64 {
			return time.Now().Unix()
		},
		internal.NewLocalCloneChangePusher(),
		internal.NewRemoteChangePusher(client),
	), nil
}

// for each option in the options slice, update the formula using those options
// the `version` arg here can often be derived from `versionutils.GetReleaseVersionOrExitGracefully()`
func (f *FormulaUpdater) Update(
	ctx context.Context,
	version *versionutils.Version,
	projectRepoOwner string,
	projectRepoName string,
	formulaOptionsList []*formula_updater_types.FormulaOptions,
) ([]formula_updater_types.FormulaStatus, error) {
	if len(formulaOptionsList) == 0 {
		return nil, nil
	}

	// our releases are tagged with a leading "v" character
	versionStrWithLeadingV := version.String()

	// we create branches/commit messages against the formula repo without the leading V
	versionStrWithNoLeadingV := versionStrWithLeadingV[1:]

	// Get version tag SHA
	// GitHub API docs: https://developer.github.com/v3/git/refs/#get-a-reference
	gitRefSha, err := f.gitClient.GetRefSha(ctx, projectRepoOwner, projectRepoName, "refs/tags/"+versionStrWithLeadingV)
	if err != nil {
		return nil, err
	}

	// Get list of release assets from GitHub
	// GitHub API docs: https://developer.github.com/v3/repos/releases/#get-a-release-by-tag-name
	releaseAssets, err := f.gitClient.GetReleaseAssetsByTag(ctx, projectRepoOwner, projectRepoName, versionStrWithLeadingV)
	if err != nil {
		return nil, err
	}

	perPlatformCliBinaryShas, err := f.getPerPlatformCliBinaryShas(releaseAssets)
	if err != nil {
		return nil, err
	}

	var formulaStatuses []formula_updater_types.FormulaStatus
	for _, formulaOptions := range formulaOptionsList {
		status := formula_updater_types.FormulaStatus{}

		// a version is not stable if it has a label, like "rc", "beta", etc.
		// in either case, silently mark it as updated and continue
		if formulaOptions.DryRun || (version.Label != "" && !formulaOptions.PublishUnstableVersion) {
			status.Updated = true
			formulaStatuses = append(formulaStatuses, status)
			continue
		}

		// Suffix branch name with random number to prevent collisions in rebuilding releases
		branchName := fmt.Sprintf("%s-%s-%d", formulaOptions.FormulaName, versionStrWithNoLeadingV, f.currentEpochTimeProvider())
		commitMessage := fmt.Sprintf("%s %s", formulaOptions.FormulaName, versionStrWithNoLeadingV)

		var changePusher formula_updater_types.ChangePusher
		if formulaOptions.PRRepoName == formulaOptions.RepoName && formulaOptions.PRRepoOwner == formulaOptions.RepoOwner {
			changePusher = f.remoteChangePusher
		} else {
			// GitHub APIs do NOT have a way to git pull --ff <remote repo>, so need to clone implementation repo locally
			// and pull remote updates.
			changePusher = f.localCloneChangePusher
		}

		err = changePusher.UpdateAndPush(ctx, versionStrWithNoLeadingV, gitRefSha, branchName, commitMessage, perPlatformCliBinaryShas, formulaOptions)
		if err != nil {
			if err == ErrAlreadyUpdated {
				status.Updated = true
			}
			status.Err = err
			formulaStatuses = append(formulaStatuses, status)
			continue
		}

		err = f.gitClient.CreatePullRequest(ctx, formulaOptions, commitMessage, branchName)
		if err == nil {
			status.Updated = true
		} else {
			status.Err = err
		}

		formulaStatuses = append(formulaStatuses, status)
	}

	return formulaStatuses, nil
}

// getGitHubSha256 extracts the sha256 strings from existing .sha256 files created as part of the build process.
// Those .sha256 files need to be located in the GitHub Release for this version.
// It returns the sha256s and any read errors encountered. It will also return ErrNoSha256sFound if any of the platform
// shas are found.
func (f *FormulaUpdater) getPerPlatformCliBinaryShas(assets []formula_updater_types.ReleaseAsset) (*formula_updater_types.PerPlatformSha256, error) {
	// Scan outputDir directory looking for any files that match the reOS regular expression as targets for extraction
	// Expect that the binaries have the platform in their name
	reOS := regexp.MustCompile("^.*(darwin|linux|windows).*.sha256$")

	shas := formula_updater_types.PerPlatformSha256{}
	for _, asset := range assets {
		s := reOS.FindStringSubmatch(asset.Name)
		if s == nil {
			continue
		}

		var err error

		switch s[1] {
		case "darwin":
			shas.DarwinSha, err = f.remoteShaGetter.GetShaFromUrl(asset.BrowserDownloadUrl)
		case "linux":
			shas.LinuxSha, err = f.remoteShaGetter.GetShaFromUrl(asset.BrowserDownloadUrl)
		case "windows":
			shas.WindowsSha, err = f.remoteShaGetter.GetShaFromUrl(asset.BrowserDownloadUrl)
		default:
			return nil, eris.Errorf("Unknown platform: '%s'", s[1])
		}

		if err != nil {
			return nil, err
		}
	}
	if shas.DarwinSha == "" && shas.LinuxSha == "" && shas.WindowsSha == "" {
		return nil, ErrNoSha256sFound
	}

	return &shas, nil
}
