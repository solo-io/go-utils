package formula_updater_types

type FormulaOptions struct {
	Name            string // Descriptive name to be used for logging and general identification
	FormulaName     string // proper formula name without file extension
	Path            string // repo relative path with file extension
	RepoOwner       string // repo owner for Formula change
	RepoName        string // repo name for Formula change
	PRRepoOwner     string // optional, empty means use RepoOwner
	PRRepoName      string // optional, empty means use RepoName
	PRBranch        string // optional, default to master
	PRDescription   string
	PRCommitName    string
	PRCommitEmail   string
	VersionRegex    string
	VersionShaRegex string
	DarwinShaRegex  string
	LinuxShaRegex   string
	WindowsShaRegex string

	// If true, open a PR even if this version is something other than a stable version. For example, "x.y.z-beta1"
	// Note that per https://docs.brew.sh/Acceptable-Formulae#stable-versions, this is not allowed for homebrew-core
	PublishUnstableVersion bool
	DryRun                 bool
}

type FormulaStatus struct {
	Name    string
	Updated bool
	Err     error
}
