package version

import (
	"github.com/pkg/errors"
	"regexp"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

var (
	zero = Version{
		Major: 0,
		Minor: 0,
		Patch: 0,
	}
)

func isGreaterThanVersion(greater, lesser *Version) bool {
	if greater.Major > lesser.Major {
		return true
	} else if greater.Major < lesser.Major {
		return false
	}

	if greater.Minor > lesser.Minor {
		return true
	} else if greater.Minor < lesser.Minor {
		return false
	}

	if greater.Patch > lesser.Patch {
		return true
	} else if greater.Patch < lesser.Patch {
		return false
	}

	return false
}

func IsGreaterThanTag(greaterTag, lesserTag string) (bool, error) {
	greaterVersion, err := ParseVersion(greaterTag)
	if err != nil {
		return false, err
	}
	lesserVersion, err := ParseVersion(lesserTag)
	if err != nil {
		return false, err
	}
	return isGreaterThanVersion(greaterVersion, lesserVersion), nil
}

func ParseVersion(tag string) (*Version, error) {
	if !MatchesRegex(tag) {
		return nil, errors.Errorf("Tag %s is not a valid semver version, must be of the form vX.Y.Z", tag)
	}
	versionString := tag[1:]
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

	version := &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
	if !isGreaterThanVersion(version, &zero) {
		return nil, errors.Errorf("Version %s is not greater than v0.0.0", tag)
	}
	return version, nil
}

func MatchesRegex(tag string) bool {
	regex := regexp.MustCompile("(v[0-9]+[.][0-9]+[.][0-9]+$)")
	return regex.MatchString(tag)
}