package api

type BuildPreparation struct {
	GithubRepo GithubRepo `json:"githubRepo"`
}

// Description of source repo
// example: https://github.com/solo-io/gloo
type GithubRepo struct {
	// Required, describes who owns the repo
	// ex: solo-io
	Owner string `json:"owner"`

	// Required
	// ex: gloo
	Repo string `json:"repo"`

	// Required
	// ex: master, feature-xyz, or someSha123etc
	Sha string `json:"sha"`

	// Optional, location to put the source files when unarchived, defaults to current directory
	// ex: passing "outputDir" would unarchive the source files into a directory like: ./outputDir/solo-io-gloo-b01c2d2/
	OutputDir string `json:"outputDir"`
}
