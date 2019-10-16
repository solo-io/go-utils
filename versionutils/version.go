package versionutils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// SemverLowerBound is the "nil" value for changelog versions
// It is not itself a valid version but it allows us to use our semver validation on the v0.0.1 edge case
const (
	SemverNilVersionValue = "v0.0.0"
	SemverMinimumVersion  = "v0.0.1"
)

var (
	InvalidSemverVersionError = func(tag string) error {
		return errors.Errorf("Tag %s is not a valid semver version, must be of the form vX.Y.Z[-rc#]", tag)
	}
)

type Version struct {
	Major            int
	Minor            int
	Patch            int
	ReleaseCandidate int
}

func NewVersion(major, minor, patch int) *Version {
	return &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
}

func NewRcVersion(major, minor, patch, rc int) *Version {
	return &Version{
		Major:            major,
		Minor:            minor,
		Patch:            patch,
		ReleaseCandidate: rc,
	}
}

func (v *Version) String() string {
	if v.ReleaseCandidate == 0 {
		return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
	}
	return fmt.Sprintf("v%d.%d.%d-rc%d", v.Major, v.Minor, v.Patch, v.ReleaseCandidate)
}

func (v *Version) IsGreaterThanOrEqualTo(lesser *Version) (bool, error) {
	if v == nil {
		return false, errors.Errorf("cannot compare versions, greater version is nil")
	}
	if lesser == nil {
		return false, errors.Errorf("cannot compare versions, lesser version is nil")
	}
	if v.ReleaseCandidate == lesser.ReleaseCandidate && v.Patch == lesser.Patch && v.Minor == lesser.Minor && v.Major == lesser.Major {
		return true, nil
	}
	return v.IsGreaterThan(lesser), nil
}

func (v *Version) IsGreaterThan(lesser *Version) bool {
	if v.Major > lesser.Major {
		return true
	} else if v.Major < lesser.Major {
		return false
	}

	if v.Minor > lesser.Minor {
		return true
	} else if v.Minor < lesser.Minor {
		return false
	}

	if v.Patch > lesser.Patch {
		return true
	} else if v.Patch < lesser.Patch {
		return false
	}

	if v.ReleaseCandidate == 0 && lesser.ReleaseCandidate > 0 {
		// special case, i.e. v1.0.0-rc1 < v1.0.0
		return true
	} else if lesser.ReleaseCandidate > 0 && v.ReleaseCandidate > lesser.ReleaseCandidate {
		return true
	}

	return false
}

func (v *Version) Equals(other *Version) bool {
	return *v == *other
}

func (v *Version) IncrementVersion(breakingChange bool) *Version {
	newMajor := v.Major
	newMinor := v.Minor
	newPatch := v.Patch
	newRc := v.ReleaseCandidate
	if v.ReleaseCandidate != 0 {
		newRc = v.ReleaseCandidate + 1
	} else if v.Major == 0 {
		if !breakingChange {
			newPatch = v.Patch + 1
		} else {
			newMinor = v.Minor + 1
			newPatch = 0
		}
	} else {
		if breakingChange {
			newMajor = v.Major + 1
			newMinor = 0
		} else {
			newMinor = v.Minor + 1
		}
		newPatch = 0
	}
	return &Version{
		Major:            newMajor,
		Minor:            newMinor,
		Patch:            newPatch,
		ReleaseCandidate: newRc,
	}
}

var (
	Zero = Version{
		Major: 0,
		Minor: 0,
		Patch: 0,
	}

	StableApiVersion = Version{
		Major: 1,
		Minor: 0,
		Patch: 0,
	}
)

func IsGreaterThanTag(greaterTag, lesserTag string) (bool, error) {
	greaterVersion, err := ParseVersion(greaterTag)
	if err != nil {
		return false, err
	}
	lesserVersion, err := ParseVersion(lesserTag)
	if err != nil {
		return false, err
	}
	return greaterVersion.IsGreaterThan(lesserVersion), nil
}

func ParseVersion(tag string) (*Version, error) {
	if !MatchesRegex(tag) {
		return nil, InvalidSemverVersionError(tag)
	}
	versionString := tag[1:]
	splitOnHyphen := strings.Split(versionString, "-")
	tagAndBuildMetadata := ""
	if len(splitOnHyphen) > 1 {
		tagAndBuildMetadata = splitOnHyphen[1]
		versionString = splitOnHyphen[0]
	}
	versionParts := strings.Split(versionString, ".")
	if len(versionParts) != 3 {
		return nil, errors.Errorf("Version %s is not a valid semver version", versionString)
	}
	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return nil, errors.Errorf("Major version %s is not valid", versionParts[0])
	}
	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return nil, errors.Errorf("Minor version %s is not valid", versionParts[1])
	}
	patch, err := strconv.Atoi(versionParts[2])
	if err != nil {
		return nil, errors.Errorf("Patch version %s is not valid", versionParts[2])
	}
	rc := 0
	if tagAndBuildMetadata != "" {
		rcString := strings.TrimPrefix(tagAndBuildMetadata, "rc")
		parsedRc, err := strconv.Atoi(rcString)
		if err != nil {
			return nil, errors.Errorf("Semver tag %s is not valid release candidate (must be 'rc' followed by int, i.e. 'rc5')", tagAndBuildMetadata)
		}
		rc = parsedRc
	}

	version := &Version{
		Major:            major,
		Minor:            minor,
		Patch:            patch,
		ReleaseCandidate: rc,
	}
	isGtEq, err := version.IsGreaterThanOrEqualTo(&Zero)
	if err != nil {
		return nil, errors.Wrapf(err, "could not compare versions")
	}
	if !isGtEq {
		return nil, errors.Errorf("Version %s is not greater than or equal to v0.0.0", tag)
	}
	return version, nil
}

func MatchesRegex(tag string) bool {
	regex := regexp.MustCompile("(v[0-9]+[.][0-9]+[.][0-9]+(-rc[0-9]+)?$)")
	return regex.MatchString(tag)
}

func GetImageVersion(version *Version) string {
	return version.String()[1:]
}
