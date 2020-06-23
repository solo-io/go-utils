package formula_updater_types

import "context"

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mocks.go

// look up git metadata; the source can be from GitHub
type GitClient interface {
	// a `ref string` is a fully qualified git ref name, like `refs/tags/1.3.0`
	GetRefSha(ctx context.Context, owner string, repo string, ref string) (string, error)

	GetReleaseAssetsByTag(ctx context.Context, owner, repo, version string) ([]ReleaseAsset, error)

	// Optionally create a pull request. We may avoid opening the PR if dry run is enabled.
	// This method will no-op and return nil if the version is a non-stable release version, and publishing non-stable versions has been disabled in the FormulaOptions
	// Expects the version string without the leading "v"
	CreatePullRequest(
		ctx context.Context,
		formulaOptions *FormulaOptions,
		commitMessage string,
		branchName string,
	) error
}

type RemoteShaGetter interface {
	// reach out to a URL and download the sha metadata there;
	// expected to be formatted like `<sha> <filename>`
	GetShaFromUrl(url string) (sha string, err error)
}

// Update the formula text, and push that change to its destination
// There are two different implementations of this: 1. when WE own the repo, and 2. when we don't (often homebrew-core).
// The difference is that in case 2, we can't do a git pull -ff through the github API, so we need to clone and update
type ChangePusher interface {
	// The error value may be ErrAlreadyUpdated if the repo has already been updated
	UpdateAndPush(
		ctx context.Context,
		version string,
		versionSha string,
		branchName string,
		commitMessage string,
		perPlatformShas *PerPlatformSha256,
		formulaOptions *FormulaOptions,
	) error
}

type Random interface {
	// replacement for golang's rand.Intn
	// used to decouple ourselves from the implementation of rand
	Intn(max int) int
}

type ReleaseAsset struct {
	Name               string
	BrowserDownloadUrl string
}

type PerPlatformSha256 struct {
	DarwinSha  string // sha256 for <ctl>-darwin binary
	LinuxSha   string // sha256 for <ctl>-linux binary
	WindowsSha string // sha256 for <ctl>-windows binary
}
