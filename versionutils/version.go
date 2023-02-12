package versionutils

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/rotisserie/eris"
)

// SemverLowerBound is the "nil" value for changelog versions
// It is not itself a valid version but it allows us to use our semver validation on the v0.0.1 edge case
const (
	SemverNilVersionValue = "v0.0.0"
	SemverMinimumVersion  = "v0.0.1"
)

// SpeciallyOrderedPrefixes are prefixes that are ordered in a special way
// This is exported so that consuming packages can specify their own special labels
// These formats take precedence over alphanumeric ordering for labels
var SpeciallyOrderedPrefixes = []string{"rc", "beta", "dev"}

var (
	InvalidSemverVersionError = func(tag string) error {
		return eris.Errorf("Tag %s is not a valid semver version, must be of the form vX.Y.Z[-<label>#]", tag)
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

func (v *Version) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

// In order, returns isGreaterThanOrEqualTo, isDeterminable, err
// isDeterminable is for incomparable versions because they have different labels, e.g. 1.0.0-foo1 vs 1.0.0-bar2
// If you want to tiebreak indeterminate comparisons using alphanumeric ordering, try MustIsGreaterThanOrEqualTo
func (v *Version) IsGreaterThanOrEqualToPtr(lesser *Version) (bool, bool, error) {
	if v == nil {
		return false, true, eris.Errorf("cannot compare versions, greater version is nil")
	}
	if lesser == nil {
		return false, true, eris.Errorf("cannot compare versions, lesser version is nil")
	}
	isGtrEq, determinable := v.IsGreaterThanOrEqualTo(*lesser)
	return isGtrEq, determinable, nil
}

// In order, returns isGreaterThanOrEqualTo, isDeterminable
// isDeterminable is for incomparable versions because they have different labels, e.g. 1.0.0-foo1 vs 1.0.0-bar2
// If you want to tiebreak indeterminate comparisons using alphanumeric ordering, try MustIsGreaterThanOrEqualTo
func (v Version) IsGreaterThanOrEqualTo(lesser Version) (bool, bool) {
	if v.Equals(&lesser) {
		return true, false
	}
	return v.IsGreaterThan(lesser)
}

// In order, returns isGreaterThanOrEqualTo, isDeterminable, err
// isDeterminable is for incomparable versions because they have different labels, e.g. 1.0.0-foo1 vs 1.0.0-bar2
// If you want to tiebreak indeterminate comparisons using alphanumeric ordering, try MustIsGreaterThan
func (v *Version) IsGreaterThanPtr(lesser *Version) (bool, bool, error) {
	if v == nil {
		return false, true, eris.Errorf("cannot compare versions, greater version is nil")
	}
	if lesser == nil {
		return false, true, eris.Errorf("cannot compare versions, lesser version is nil")
	}
	isGreater, determinable := v.IsGreaterThan(*lesser)
	return isGreater, determinable, nil
}

// In order, returns isGreaterThanOrEqualTo, isDeterminable
// isDeterminable is for incomparable versions because they have different labels, e.g. 1.0.0-foo1 vs 1.0.0-bar2
// If you want to tiebreak indeterminate comparisons using alphanumeric ordering, try MustIsGreaterThan
func (v Version) IsGreaterThan(lesser Version) (bool, bool) {
	if v.Major > lesser.Major {
		return true, true
	} else if v.Major < lesser.Major {
		return false, true
	}

	if v.Minor > lesser.Minor {
		return true, true
	} else if v.Minor < lesser.Minor {
		return false, true
	}

	if v.Patch > lesser.Patch {
		return true, true
	} else if v.Patch < lesser.Patch {
		return false, true
	}

	if len(v.Label) == 0 && lesser.Label != "" {
		return true, true
	} else if len(v.Label) > 0 && len(lesser.Label) == 0 {
		return false, true
	}

	if v.Label == lesser.Label {
		if v.LabelVersion > lesser.LabelVersion {
			return true, true
		}
		// is determinabley not greater than
		return false, true

	}

	// impose additional ordering based on our special labels
	for _, label := range SpeciallyOrderedPrefixes {

		if v.Label == label {
			// we know that they arent the same so we can return immediately
			return true, true
		}
		if lesser.Label == label {
			return false, true
		}
	}

	return false, false
}

// In order, returns isGreaterThanOrEqualTo, isDeterminable
// labelOrder specifies tie-break order for labels
// e.g. labelOrder = [ beta, alpha, predev ], then 1.7.0-beta11 > 1.7.0-alpha5 > 1.7.0-predev9
// isDeterminable is for incomporable versions because they have different labels not specified in labelOrder
func (v Version) IsGreaterThanWithLabelOrder(lesser Version, labelOrder []string) (bool, bool) {

	greaterThan, determinable := v.IsGreaterThan(lesser)
	if determinable {
		return greaterThan, determinable
	}
	vIndex, lIndex := -1, -1
	for i, lbl := range labelOrder {
		if lbl == v.Label {
			vIndex = i
		}
		if lbl == lesser.Label {
			lIndex = i
		}
	}
	if vIndex == -1 || lIndex == -1 {
		return false, false
	}

	return vIndex < lIndex, true

}

// for incomparable versions, default to alphanumeric sort on label
// e.g. 1.0.0-foo1 > 1.0.0-bar2
func (v Version) MustIsGreaterThanOrEqualTo(lesser Version) bool {
	if v.Equals(&lesser) {
		return true
	}
	return v.MustIsGreaterThan(lesser)
}

// MustIsGreaterThan is like IsGreaterThan, but for incomparable versions, default to alphanumeric sort on label
// e.g. 1.0.0-foo1 > 1.0.0-bar2
func (v Version) MustIsGreaterThan(lesser Version) bool {
	isGtr, determinable := v.IsGreaterThan(lesser)
	if determinable {
		return isGtr
	}
	// if we can't compare versions (i.e., different labels not in our special set)
	// then default to alphanumeric sort

	return v.Label > lesser.Label
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

func Zero() Version {
	return Version{
		Major: 0,
		Minor: 0,
		Patch: 0,
	}
}

func StableApiVersion() Version {
	return Version{
		Major: 1,
		Minor: 0,
		Patch: 0,
	}
}

func IsGreaterThanTag(greaterTag, lesserTag string) (bool, bool, error) {
	greaterVersion, err := ParseVersion(greaterTag)
	if err != nil {
		return false, false, err
	}
	lesserVersion, err := ParseVersion(lesserTag)
	if err != nil {
		return false, false, err
	}
	return greaterVersion.IsGreaterThanPtr(lesserVersion)
}

func IsGreaterThanTagWithLabelOrder(greaterTag, lesserTag string, labelOrder []string) (bool, bool, error) {
	greaterVersion, err := ParseVersion(greaterTag)
	if err != nil {
		return false, false, err
	}
	lesserVersion, err := ParseVersion(lesserTag)
	if err != nil {
		return false, false, err
	}
	isGreaterThan, determinable := greaterVersion.IsGreaterThanWithLabelOrder(*lesserVersion, labelOrder)
	return isGreaterThan, determinable, nil
}

func ParseVersion(tag string) (*Version, error) {
	if !MatchesRegex(tag) {
		return nil, InvalidSemverVersionError(tag)
	}

	versionString := tag[1:]
	splitOnHyphen := strings.SplitN(versionString, "-", 2)
	labelAndVersion := ""
	if len(splitOnHyphen) > 1 {
		labelAndVersion = splitOnHyphen[1]
		versionString = splitOnHyphen[0]
	}
	versionParts := strings.Split(versionString, ".")
	if len(versionParts) != 3 {
		return nil, eris.Errorf("Version %s is not a valid semver version", versionString)
	}
	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return nil, eris.Errorf("Major version %s is not valid", versionParts[0])
	}
	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return nil, eris.Errorf("Minor version %s is not valid", versionParts[1])
	}
	patch, err := strconv.Atoi(versionParts[2])
	if err != nil {
		return nil, eris.Errorf("Patch version %s is not valid", versionParts[2])
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

	isGtEq, _ := version.IsGreaterThanOrEqualTo(Zero())
	if !isGtEq {
		return nil, eris.Errorf("Version %s is not greater than or equal to v0.0.0", tag)
	}
	return version, nil
}

func parseLabelVersion(labelAndVersion string) (string, int, error) {
	numsRegex := regexp.MustCompile("[0-9]+")
	numberGroups := numsRegex.FindAllString(labelAndVersion, -1)
	if len(numberGroups) == 0 {
		// valid label with no version, eg "wasm"
		return labelAndVersion, 0, nil
	}

	label := numsRegex.ReplaceAllString(labelAndVersion, "")
	labelVersionToParse := numberGroups[0]
	labelVersion, err := strconv.Atoi(labelVersionToParse)
	if err != nil {
		return "", 0, errors.Wrapf(err, "invalid label version %s", labelVersionToParse)
	}

	return label, labelVersion, nil
}

func MatchesRegex(tag string) bool {
	regex := regexp.MustCompile("(v[0-9]+[.][0-9]+[.][0-9]+(-[a-z]+)*(-[a-z]+[0-9]*)?$)")
	return regex.MatchString(tag)
}

func GetImageVersion(version *Version) string {
	return version.String()[1:]
}

func Index(versions []Version, v Version) int {
	for idx, ver := range versions {
		if v == ver {
			return idx
		}
	}
	return -1
}

func IndexPtr(versions []*Version, v Version) int {
	for idx, ver := range versions {
		if v == *ver {
			return idx
		}
	}
	return -1
}
