package versionutils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/solo-io/go-utils/versionutils/dep"

	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
)

const (
	gopkgToml     = "Gopkg.toml"
	constraint    = "constraint"
	override      = "override"
	nameConst     = "name"
	versionConst  = "version"
	revisionConst = "revision"
	branchConst   = "branch"
)

var (
	UnableToFindVersionInTomlError = func(pkgName string) error {
		return fmt.Errorf("unable to find version for %s in toml", pkgName)
	}
)

func PinGitVersion(relativeRepoDir string, version string) error {
	tag := GetTag(version)
	cmd := exec.Command("git", "checkout", tag)
	cmd.Dir = relativeRepoDir
	buf := &bytes.Buffer{}
	out := io.MultiWriter(buf, os.Stdout)
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "%v failed: %s", cmd.Args, buf.String())
	}
	return nil
}

func GetGitVersion(relativeRepoDir string) (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--dirty")
	cmd.Dir = relativeRepoDir
	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrapf(err, "%v failed: %s", cmd.Args, output)
	}
	return strings.TrimSpace(string(output)), nil
}

func GetTag(version string) string {
	if strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}

func GetVersionFromTag(shouldBeAVersion string) (string, error) {
	definiteTag := GetTag(shouldBeAVersion)
	version, err := ParseVersion(definiteTag)
	if err != nil {
		return "", err
	}
	versionString := version.String()
	return versionString[1:], nil
}

// Deprecated: Use GetTomlVersion instead
func GetVersion(pkgName string, tomlTree []*toml.Tree) (string, error) {
	for _, v := range tomlTree {
		if version, found := getVersionInfoFromTree(v, pkgName); found {
			return version.Version, nil
		}
	}
	return "", UnableToFindVersionInTomlError(pkgName)
}

// Deprecated: use GetDependencyVersionInfo
func GetTomlVersion(pkgName string, toml *TomlWrapper) (string, error) {
	for _, v := range toml.Overrides {
		if version, found := getVersionInfoFromTree(v, pkgName); found {
			return version.Version, nil
		}
	}
	for _, v := range toml.Constraints {
		if version, found := getVersionInfoFromTree(v, pkgName); found {
			return version.Version, nil
		}
	}
	return "", UnableToFindVersionInTomlError(pkgName)
}

// Returns the version of the given package together with the type of version identifier, i.e. revision, version, branch.
func GetDependencyVersionInfo(pkgName string, toml *TomlWrapper) (*dep.VersionInfo, error) {
	for _, v := range toml.Overrides {
		if version, found := getVersionInfoFromTree(v, pkgName); found {
			return version, nil
		}
	}
	for _, v := range toml.Constraints {
		if version, found := getVersionInfoFromTree(v, pkgName); found {
			return version, nil
		}
	}
	return nil, UnableToFindVersionInTomlError(pkgName)
}

// Deprecated: Use ParseFullToml instead
func ParseToml() ([]*toml.Tree, error) {
	return ParseTomlFromDir("")
}

// Deprecated: Use ParseFullTomlFromDir instead
func ParseTomlFromDir(relativeDir string) ([]*toml.Tree, error) {
	return parseTomlFromDir(relativeDir, constraint)
}

// Deprecated: Use ParseFullToml instead
func ParseTomlOverrides() ([]*toml.Tree, error) {
	return ParseTomlOverridesFromDir("")
}

// Deprecated: Use ParseFullTomlFromDir instead
func ParseTomlOverridesFromDir(relativeDir string) ([]*toml.Tree, error) {
	return parseTomlFromDir(relativeDir, override)
}

func parseTomlFromDir(relativeDir, configType string) ([]*toml.Tree, error) {
	config, err := toml.LoadFile(filepath.Join(relativeDir, gopkgToml))
	if err != nil {
		return nil, err
	}

	tomlTree := config.Get(configType)

	switch typedTree := tomlTree.(type) {
	case []*toml.Tree:
		return typedTree, nil
	default:
		return nil, fmt.Errorf("unable to parse toml tree")
	}
}

type TomlWrapper struct {
	Overrides   []*toml.Tree
	Constraints []*toml.Tree
}

func ParseFullTomlFromDir(relativeDir string) (*TomlWrapper, error) {
	overrides, err := ParseTomlOverridesFromDir(relativeDir)
	if err != nil {
		return nil, err
	}
	constraints, err := ParseTomlFromDir(relativeDir)
	if err != nil {
		return nil, err
	}
	return &TomlWrapper{
		Constraints: constraints,
		Overrides:   overrides,
	}, nil
}

func ParseFullToml() (*TomlWrapper, error) {
	return ParseFullTomlFromDir("")
}

func getVersionInfoFromTree(tomlTree *toml.Tree, pkgName string) (info *dep.VersionInfo, found bool) {
	isEmpty := func(node *toml.Tree, key string) bool {
		return node.Get(key) == nil || node.Get(key) == ""
	}

	if tomlTree.Get(nameConst) != pkgName {
		return nil, false
	}

	switch {
	case !isEmpty(tomlTree, versionConst):
		return &dep.VersionInfo{
			Version: tomlTree.Get(versionConst).(string),
			Type:    dep.Version,
		}, true
	case !isEmpty(tomlTree, revisionConst):
		return &dep.VersionInfo{
			Version: tomlTree.Get(revisionConst).(string),
			Type:    dep.Revision,
		}, true
	case !isEmpty(tomlTree, branchConst):
		return &dep.VersionInfo{
			Version: tomlTree.Get(branchConst).(string),
			Type:    dep.Branch,
		}, true
	}
	return nil, false
}
