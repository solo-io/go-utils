package botutils

import (
	"github.com/google/go-github/github"

	"github.com/palantir/go-githubapp/githubapp"
)

type GHClient struct {
	ClientCreator githubapp.ClientCreator
	Token         string
}

func (h *GHClient) getClient(installationID int64) (*github.Client, error) {
	if h.Token != "" {
		return h.ClientCreator.NewTokenClient(h.Token)
	}
	return h.ClientCreator.NewInstallationClient(installationID)
}
