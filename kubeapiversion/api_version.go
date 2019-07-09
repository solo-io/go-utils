package kubeapiversion

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/solo-io/go-utils/errors"
)

var (
	apiVersionRegexp *regexp.Regexp

	MalformedVersionError = func(version string) error {
		return errors.Errorf("Failed to parse kubernetes api version from %v", version)
	}

	InvalidMajorVersionError = errors.New("Major version cannot be zero")

	InvalidPrereleaseVersionError = errors.New("Prerelease version cannot be zero")
)

type PrereleaseModifier int

const (
	Alpha PrereleaseModifier = iota + 1
	Beta
	GA
)

func (m PrereleaseModifier) String() string {
	switch m {
	case Alpha:
		return "alpha"
	case Beta:
		return "beta"
	default:
		return ""
	}
}

func init() {
	apiVersionRegexp = regexp.MustCompile(`^v([0-9]+)((alpha|beta)([0-9]+))?$`)
}

type ApiVersion interface {
	Major() int
	Prerelease() int
	PrereleaseModifier() PrereleaseModifier
	String() string
	GreaterThan(other ApiVersion) bool
	LessThan(other ApiVersion) bool
	Equal(other ApiVersion) bool
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

func (v *apiVersion) GreaterThan(other ApiVersion) bool {
	if v.Major() < other.Major() {
		return false
	}

	if v.Major() == other.Major() {
		if v.PrereleaseModifier() < other.PrereleaseModifier() {
			return false
		}

		if v.PrereleaseModifier() == other.PrereleaseModifier() {
			if v.Prerelease() <= other.Prerelease() {
				return false
			}
		}
	}

	return true
}

func (v *apiVersion) LessThan(other ApiVersion) bool {
	if v.Major() > other.Major() {
		return false
	}

	if v.Major() == other.Major() {
		if v.PrereleaseModifier() > other.PrereleaseModifier() {
			return false
		}

		if v.PrereleaseModifier() == other.PrereleaseModifier() {
			if v.Prerelease() >= other.Prerelease() {
				return false
			}
		}
	}

	return true
}

func (v *apiVersion) Equal(other ApiVersion) bool {
	return v.String() == other.String()
}
