package kubeutils

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/solo-io/go-utils/errors"
)

type PrereleaseModifier int

var (
	apiVersionRegexp *regexp.Regexp

	MalformedVersionError = func(version string) error {
		return errors.Errorf("Failed to parse kubernetes api version from %v", version)
	}

	InvalidMajorVersionError = errors.Errorf("Major version cannot be zero")

	InvalidPrereleaseVersionError = errors.Errorf("Prerelease version cannot be zero")
)

const (
	Alpha PrereleaseModifier = iota + 1
	Beta
	GA
)

func init() {
	apiVersionRegexp = regexp.MustCompile(`^v([0-9]+)((alpha|beta)([0-9]+))?$`)
}

type ApiVersion interface {
	Major() int
	Prerelease() int
	PrereleaseModifier() PrereleaseModifier
	String() string
	GreaterThan(version ApiVersion) bool
	LessThan(version ApiVersion) bool
	Equal(version ApiVersion) bool
}

type apiVersion struct {
	major, prerelease int
	modifier          PrereleaseModifier
}

func ParseApiVersion(version string) (ApiVersion, error) {
	matches := apiVersionRegexp.FindStringSubmatch(version)
	if matches == nil {
		return nil, MalformedVersionError(version)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, err
	}
	if major == 0 {
		return nil, InvalidMajorVersionError
	}

	// Prerelease version info is optional
	var prerelease int
	if matches[3] != "" && matches[4] != "" {
		prerelease, err = strconv.Atoi(matches[4])
		if err != nil {
			return nil, err
		}
		if prerelease == 0 {
			return nil, InvalidPrereleaseVersionError
		}
	}

	var modifier PrereleaseModifier
	switch matches[3] {
	case "alpha":
		modifier = Alpha
	case "beta":
		modifier = Beta
	default:
		modifier = GA
	}

	return &apiVersion{
		major:      major,
		prerelease: prerelease,
		modifier:   modifier,
	}, nil
}

func NewApiVersion(major, prerelease int, prereleaseModifier PrereleaseModifier) ApiVersion {
	return &apiVersion{
		major:      major,
		prerelease: prerelease,
		modifier:   prereleaseModifier,
	}
}

func (v *apiVersion) Major() int {
	return v.major
}

func (v *apiVersion) Prerelease() int {
	return v.prerelease
}

func (v *apiVersion) PrereleaseModifier() PrereleaseModifier {
	return v.modifier
}

func (v *apiVersion) String() string {
	sb := strings.Builder{}
	sb.WriteString("v")
	sb.WriteString(strconv.Itoa(v.Major()))

	switch v.PrereleaseModifier() {
	case Alpha:
		sb.WriteString("alpha")
		sb.WriteString(strconv.Itoa(v.Prerelease()))
	case Beta:
		sb.WriteString("beta")
		sb.WriteString(strconv.Itoa(v.Prerelease()))
	}

	return sb.String()
}

func (v *apiVersion) GreaterThan(version ApiVersion) bool {
	if v.Major() < version.Major() {
		return false
	}

	if v.Major() == version.Major() {
		if v.PrereleaseModifier() < version.PrereleaseModifier() {
			return false
		}

		if v.PrereleaseModifier() == version.PrereleaseModifier() {
			if v.Prerelease() <= version.Prerelease() {
				return false
			}
		}
	}

	return true
}

func (v *apiVersion) LessThan(version ApiVersion) bool {
	if v.Major() > version.Major() {
		return false
	}

	if v.Major() == version.Major() {
		if v.PrereleaseModifier() > version.PrereleaseModifier() {
			return false
		}

		if v.PrereleaseModifier() == version.PrereleaseModifier() {
			if v.Prerelease() >= version.Prerelease() {
				return false
			}
		}
	}

	return true
}

func (v *apiVersion) Equal(version ApiVersion) bool {
	return v.Major() == version.Major() && v.PrereleaseModifier() == v.PrereleaseModifier() && v.Prerelease() == v.Prerelease()
}
