package changelogutils

import (
	"github.com/solo-io/go-utils/versionutils"

	"github.com/pkg/errors"
)

func validVersion(version, proposedVersion, latestTag string) (string, error) {
	var newProposedVersion string

	if !versionutils.MatchesRegex(version) {
		return "", errors.Errorf("Directory name %s is not valid, must be of the form 'vX.Y.Z'", version)
	}
	greaterThan, err := versionutils.IsGreaterThanTag(version, latestTag)
	if err != nil {
		return "", err
	}
	if greaterThan {
		if proposedVersion != "" {
			return "", errors.Errorf("Versions %s and %s are both greater than latest tag", version, proposedVersion)
		}
		newProposedVersion = version
	}
	return newProposedVersion, nil

}
