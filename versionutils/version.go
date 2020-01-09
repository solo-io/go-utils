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
	UnableToCompareVersionError = func(v1 string, v2 string) error {
		return errors.Errorf("Unable to compare version %s to %s.", v1, v2)
	}
)

type Version struct {
	Major int
	Minor int
	Patch int

	// optional to support a version like "1.0.0-rc1", where "rc" is the label and "1" is the label version
	// for comparisons:
	//  - "1.0.0-rc1" is greater than "0.X.Y" and less than "1.0.0"
	//  - "1.0.0-rc5" is greater than "1.0.0-rc1"
	//  - "1.0.0-aX" is not greater than or less than "1.0.0-bY", except by convention
	Label        string
	LabelVersion int
}

func NewVersion(major, minor, patch int, label string, labelVersion int) *Version {
	return &Version{
		Major:        major,
		Minor:        minor,
		Patch:        patch,
		Label:        label,
		LabelVersion: labelVersion,
	}
}

func (v *Version) String() string {
	if v.LabelVersion == 0 {
		return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
	}
	return fmt.Sprintf("v%d.%d.%d-%s%d", v.Major, v.Minor, v.Patch, v.Label, v.LabelVersion)
}

// users may want to handle UnableToCompareVersionError
// this is for versions that cannot be compared because they have different labels, e.g. 1.0.0-foo1 vs 1.0.0-bar2
func (v *Version) IsGreaterThanOrEqualTo(lesser *Version) (bool, error) {
	if v == nil {
		return false, errors.Errorf("cannot compare versions, greater version is nil")
	}
	if lesser == nil {
		return false, errors.Errorf("cannot compare versions, lesser version is nil")
	}
	if v.Equals(lesser) {
		return true, nil
	}
	return v.IsGreaterThan(lesser)
}

// users may want to handle UnableToCompareVersionError
// this is for versions that cannot be compared because they have different labels, e.g. 1.0.0-foo1 vs 1.0.0-bar2
func (v *Version) IsGreaterThan(lesser *Version) (bool, error) {
	if v.Major > lesser.Major {
		return true, nil
	} else if v.Major < lesser.Major {
		return false, nil
	}

	if v.Minor > lesser.Minor {
		return true, nil
	} else if v.Minor < lesser.Minor {
		return false, nil
	}

	if v.Patch > lesser.Patch {
		return true, nil
	} else if v.Patch < lesser.Patch {
		return false, nil
	}

	if len(v.Label) == 0 && len(lesser.Label) > 0 {
		return true, nil
	} else if len(v.Label) > 0 && len(lesser.Label) == 0 {
		return false, nil
	}

	if v.Label != lesser.Label {
		return false, UnableToCompareVersionError(v.String(), lesser.String())
	}

	if v.LabelVersion > lesser.LabelVersion {
		return true, nil
	} else if v.LabelVersion < lesser.LabelVersion {
		return false, nil
	}

	return false, nil
}

func (v *Version) Equals(other *Version) bool {
	return *v == *other
}

func (v *Version) IncrementVersion(breakingChange, newFeature bool) *Version {
	newMajor := v.Major
	newMinor := v.Minor
	newPatch := v.Patch
	newLabelVersion := v.LabelVersion
	if v.LabelVersion != 0 {
		newLabelVersion = v.LabelVersion + 1
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
			newPatch = 0
		} else if newFeature {
			newMinor = v.Minor + 1
			newPatch = 0
		} else {
			newPatch = v.Patch + 1
		}
	}
	return &Version{
		Major:        newMajor,
		Minor:        newMinor,
		Patch:        newPatch,
		Label:        v.Label,
		LabelVersion: newLabelVersion,
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
	return greaterVersion.IsGreaterThan(lesserVersion)
}

func ParseVersion(tag string) (*Version, error) {
	if !MatchesRegex(tag) {
		return nil, InvalidSemverVersionError(tag)
	}
	versionString := tag[1:]
	splitOnHyphen := strings.Split(versionString, "-")
	labelAndVersion := ""
	if len(splitOnHyphen) > 1 {
		labelAndVersion = splitOnHyphen[1]
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

	version := &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}

	if labelAndVersion != "" {
		label, labelVersion, err := parseLabelVersion(labelAndVersion)
		if err != nil {
			return nil, err
		}
		version.Label = label
		version.LabelVersion = labelVersion
	}

	isGtEq, err := version.IsGreaterThanOrEqualTo(&Zero)
	if err != nil {
		return nil, err
	}
	if !isGtEq {
		return nil, errors.Errorf("Version %s is not greater than or equal to v0.0.0", tag)
	}
	return version, nil
}

func parseLabelVersion(labelAndVersion string) (string, int, error) {
	regex := regexp.MustCompile("([a-z]+)([0-9]+)")
	// should be like ["foo1", "foo", "1"]
	matches := regex.FindStringSubmatch(labelAndVersion)
	if len(matches) != 3 {
		return "", 0, errors.Errorf("invalid label and version %s", labelAndVersion)
	}
	label := matches[1]
	labelVersionToParse := matches[2]
	labelVersion, err := strconv.Atoi(labelVersionToParse)
	if err != nil {
		return "", 0, errors.Wrapf(err, "invalid label version %s", labelVersionToParse)
	}
	return label, labelVersion, nil
}

func MatchesRegex(tag string) bool {
	regex := regexp.MustCompile("(v[0-9]+[.][0-9]+[.][0-9]+(-[a-z]+[0-9]+)?$)")
	return regex.MatchString(tag)
}

func GetImageVersion(version *Version) string {
	return version.String()[1:]
}
